package pkg

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"

	"github.com/E4-Computer-Engineering/nvme_exporter/pkg/utils"
)

// GetDevices queries the devices list through the shell
// and returns an array of JSON results with the devices data.
// This function handles both old flat structure and new nested structure
// of nvme-cli JSON output.
func GetDevices() []gjson.Result {
	// Check validation state before attempting to query devices
	if validationChecker != nil && !validationChecker() {
		if scrapeFailureIncrementer != nil {
			scrapeFailureIncrementer()
		}
		log.Printf("Skipping device query due to validation failure")
		return []gjson.Result{}
	}

	devicesJSON, err := utils.ExecuteJSONCommand("nvme", "list", "-o", "json")
	if err != nil {
		log.Printf("Error running nvme list -o json: %s\n", err)
		if scrapeFailureIncrementer != nil {
			scrapeFailureIncrementer()
		}
		return []gjson.Result{}
	}

	devices := devicesJSON.Get("Devices").Array()
	if len(devices) == 0 {
		return []gjson.Result{}
	}

	// Check if we have the new nested structure (with Subsystems)
	// or the old flat structure (with DevicePath)
	firstDevice := devices[0]
	if firstDevice.Get("Subsystems").Exists() {
		// New nested structure - flatten it
		return flattenNewStructure(devices)
	}

	// Old flat structure - return as is
	return devices
}

// flattenNewStructure converts the new nested nvme-cli JSON structure
// to a flat structure compatible with the rest of the code.
func flattenNewStructure(devices []gjson.Result) []gjson.Result {
	var flattened []gjson.Result

	for _, device := range devices {
		subsystems := device.Get("Subsystems").Array()
		for _, subsystem := range subsystems {
			controllers := subsystem.Get("Controllers").Array()
			for _, controller := range controllers {
				serialNumber := controller.Get("SerialNumber").String()
				modelNumber := controller.Get("ModelNumber").String()
				firmware := controller.Get("Firmware").String()

				namespaces := controller.Get("Namespaces").Array()
				for _, namespace := range namespaces {
					namespaceName := namespace.Get("NameSpace").String()
					generic := namespace.Get("Generic").String()

					// Construct a flat JSON object compatible with old structure
					flatJSON := map[string]interface{}{
						"DevicePath":   "/dev/" + namespaceName,
						"GenericPath":  generic,
						"Firmware":     firmware,
						"ModelNumber":  modelNumber,
						"SerialNumber": serialNumber,
						"NameSpace":    namespace.Get("NSID").Int(),
						"UsedBytes":    namespace.Get("UsedBytes").Int(),
						"MaximumLBA":   namespace.Get("MaximumLBA").Int(),
						"PhysicalSize": namespace.Get("PhysicalSize").Int(),
						"SectorSize":   namespace.Get("SectorSize").Int(),
					}

					// Convert map to JSON string and parse it as gjson.Result
					jsonStr := utils.MapToJSONString(flatJSON)
					flattened = append(flattened, gjson.Parse(jsonStr))
				}
			}
		}
	}

	return flattened
}

// MetricCollector is the interface implemented by the objects contained
// in the CompositeCollector field.
//
// We could have the collectors implement prometheus.Collector directly, but then
// we wound unnecessarily have to call GetDevices more than once.
//
// Here the device data is injected in CollectMetrics, so that we can call
// GetDevice once in CompositeCollector.Collect
// (it is a shell function, so every call can potentially be "expensive").
type MetricCollector interface {
	// Describe is the same as prometheus.Collector.Describe
	Describe(descChan chan<- *prometheus.Desc)

	// CollectMetrics does what prometheus.Collector.Collect does,
	// but needs the device JSON data to prevent calling GetDevice
	// multiple times
	CollectMetrics(metricChan chan<- prometheus.Metric, device gjson.Result)
}

// InfoMetricCollector implements MetricCollector and sends info metrics.
type InfoMetricCollector struct {
	// InfoMetricProviders is the list of providers for the info metric collector
	InfoMetricProviders []MetricProvider
}

// NewInfoMetricCollector initializes and returns a new InfoMetricCollector object.
func NewInfoMetricCollector(providers []MetricProvider) *InfoMetricCollector {
	return &InfoMetricCollector{InfoMetricProviders: providers}
}

// Describe sends all prometheus.Desc pointers through the channel.
func (ic *InfoMetricCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, infoProvider := range ic.InfoMetricProviders {
		ch <- infoProvider.Desc
	}
}

// CollectMetrics gets the devices data and sends all info metrics through the channel.
func (ic *InfoMetricCollector) CollectMetrics(ch chan<- prometheus.Metric, device gjson.Result) {
	devicePath := device.Get("DevicePath").String()
	genericPath := device.Get("GenericPath").String()
	firmware := device.Get("Firmware").String()
	modelNumber := device.Get("ModelNumber").String()
	serialNumber := device.Get("SerialNumber").String()

	for _, infoProvider := range ic.InfoMetricProviders {
		// Fetching the metric object is delegated to the provider
		metric := infoProvider.GetMetric(
			device,
			devicePath,
			genericPath,
			firmware,
			modelNumber,
			serialNumber,
		)
		// Only send metric if it's not nil (handles cases where data is unavailable)
		if metric != nil {
			ch <- metric
		}
	}
}

// InfoMetricCollector implements MetricCollector and sends smart log metrics.
type LogMetricCollector struct {
	// LogMetricProviders is the list of providers for the log metric collector
	LogMetricProviders []MetricProvider

	// getData receives the devicePath and gets the log JSON data
	getData func(string) gjson.Result
}

// NewLogMetricCollector initializes and returns a new LogMetricCollector object.
func NewLogMetricCollector(providers []MetricProvider, getData func(string) gjson.Result) *LogMetricCollector {
	return &LogMetricCollector{
		LogMetricProviders: providers,
		getData:            getData,
	}
}

// Describe sends all prometheus.Desc pointers through the channel.
func (lc *LogMetricCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, logProvider := range lc.LogMetricProviders {
		ch <- logProvider.Desc
	}
}

// Collect gets the smart log data and sends all log metrics through the channel.
func (lc *LogMetricCollector) CollectMetrics(ch chan<- prometheus.Metric, device gjson.Result) {
	devicePath := device.Get("DevicePath").String()

	jsonData := lc.getData(devicePath)

	// If getData returns invalid data (e.g., OCP not supported), skip this collector
	if !jsonData.Exists() {
		return
	}

	for _, logProvider := range lc.LogMetricProviders {
		// Fetching the metric object is delegated to the provider
		metric := logProvider.GetMetric(jsonData, devicePath)
		// Only send metric if it's not nil (handles cases where data is unavailable)
		if metric != nil {
			ch <- metric
		}
	}
}

// CompositeCollector implements prometheus.Collector interface,
// wrapping a slice of other MetricCollector objects.
type CompositeCollector struct {
	// collectors holds a simple list of MetricCollector objects
	collectors []MetricCollector
}

// NewCompositeCollector initializes and returns a new CompositeCollector object.
func NewCompositeCollector(collectors []MetricCollector) *CompositeCollector {
	return &CompositeCollector{collectors: collectors}
}

// Describe calls Describe on every collector in cc.collectors.
func (cc *CompositeCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, collector := range cc.collectors {
		collector.Describe(ch)
	}
}

// Collect calls Collect on every collector in cc.collectors.
func (cc *CompositeCollector) Collect(ch chan<- prometheus.Metric) {
	devices := GetDevices()

	for _, device := range devices {
		for _, collector := range cc.collectors {
			collector.CollectMetrics(ch, device)
		}
	}
}

// SetValidationChecker allows injecting a validation check function
// to determine if scrapes should proceed
type ValidationChecker func() bool

var validationChecker ValidationChecker

// SetValidationChecker sets the global validation checker
func SetValidationChecker(checker ValidationChecker) {
	validationChecker = checker
}

// IncrementScrapeFailure is called when a scrape fails due to validation errors
type ScrapeFailureIncrementer func()

var scrapeFailureIncrementer ScrapeFailureIncrementer

// SetScrapeFailureIncrementer sets the global scrape failure incrementer
func SetScrapeFailureIncrementer(incrementer ScrapeFailureIncrementer) {
	scrapeFailureIncrementer = incrementer
}
