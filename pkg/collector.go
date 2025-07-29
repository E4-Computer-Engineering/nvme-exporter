package pkg

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"

	"github.com/E4-Computer-Engineering/nvme_exporter/pkg/utils"
)

// GetDevices queries the devices list through the shell
// and returns an array of JSON results with the devices data.
func GetDevices() []gjson.Result {
	devicesJSON, err := utils.ExecuteJSONCommand("nvme", "list", "-o", "json")
	if err != nil {
		log.Printf("Error running nvme list -o json: %s\n", err)
	}

	return devicesJSON.Get("Devices").Array()
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
		ch <- metric
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
	for _, logProvider := range lc.LogMetricProviders {
		// Fetching the metric object is delegated to the provider
		metric := logProvider.GetMetric(jsonData, devicePath)
		ch <- metric
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
