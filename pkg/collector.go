package pkg

import (
	"log"

	"github.com/E4-Computer-Engineering/nvme_exporter/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
)

// NvmeCollector implements prometheus.Collector interface.
// Metric provider objects are separated in three different fields
type NvmeCollector struct {
	OcpEnabled            bool
	InfoMetricProviders   []InfoMetricProvider
	LogMetricProviders    []LogMetricProvider
	OcpLogMetricProviders []LogMetricProvider
}

// Describe now is a compact method that iterates
// through the MetricProvider slices, and sends all prometheus.Desc
// object through the channel
func (c *NvmeCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, infoProvider := range c.InfoMetricProviders {
		ch <- infoProvider.Desc
	}

	for _, logProvider := range c.LogMetricProviders {
		ch <- logProvider.Desc
	}

	for _, ocpLogProvider := range c.OcpLogMetricProviders {
		ch <- ocpLogProvider.Desc
	}
}

// Collect gets the devices list and sends all the needed
// metrics through the provided channel
func (c *NvmeCollector) Collect(ch chan<- prometheus.Metric) {
	devices := c.getDevices()

	for _, device := range devices {
		c.sendInfoMetrics(ch, device)
		c.sendSmartLogMetrics(ch, device)

		if c.OcpEnabled {
			c.sendOcpSmartLogMetrics(ch, device)
		}
	}
}

// getDevices queries the devices list through the shell
// and returns an array of JSON results with the devices data
func (c *NvmeCollector) getDevices() []gjson.Result {
	shell := utils.NewShell(utils.WithValidators(gjson.Valid))
	nvmeDeviceCmd, err := shell.Run("nvme", "list", "-o", "json")
	if err != nil {
		log.Printf("Error running nvme list -o json: %s\n", err)
	}

	return gjson.Get(string(nvmeDeviceCmd), "Devices").Array()
}

// sendInfoMetrics gets the info metric from all InfoMetricProvider
// in c.InfoMetricProviders, and sends them through the channel
func (c *NvmeCollector) sendInfoMetrics(ch chan<- prometheus.Metric, device gjson.Result) {
	devicePath := device.Get("DevicePath").String()
	genericPath := device.Get("GenericPath").String()
	firmware := device.Get("Firmware").String()
	modelNumber := device.Get("ModelNumber").String()
	serialNumber := device.Get("SerialNumber").String()

	for _, infoProvider := range c.InfoMetricProviders {
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

// sendSmartLogMetrics queries the shell for smart-log data,
// gets the metrics from each LogMetricProvider in c.LogMetricProviders
// and sends them through the channel
func (c *NvmeCollector) sendSmartLogMetrics(ch chan<- prometheus.Metric, device gjson.Result) {
	shell := utils.NewShell(utils.WithValidators(gjson.Valid))
	smartLog, err := shell.Run("nvme", "smart-log", device.String(), "-o", "json")
	if err != nil {
		log.Printf("Error running smart-log %s -o json: %s\n", device.String(), err)
	}

	jsonLog := gjson.Parse(string(smartLog))

	for _, logProvider := range c.LogMetricProviders {
		// Fetching the metric object is delegated to the provider
		metric := logProvider.GetMetric(jsonLog, device.String())
		ch <- metric
	}
}

// sendOcpSmartLogMetrics queries the shell for ocp smart-log data,
// gets the metrics from each OcpLogMetricProvider in c.OcpLogMetricProviders
// and sends them through the channel
func (c *NvmeCollector) sendOcpSmartLogMetrics(ch chan<- prometheus.Metric, device gjson.Result) {
	shell := utils.NewShell(utils.WithValidators(gjson.Valid))
	ocpSmartLog, err := shell.Run("nvme", "ocp", "smart-add-log", device.String(), "-o", "json")
	if err != nil {
		log.Printf("Error running smart-add-log %s -o json: %s\n", device.String(), err)
	}

	jsonLog := gjson.Parse(string(ocpSmartLog))

	for _, ocpLogProvider := range c.OcpLogMetricProviders {
		// Fetching the metric object is delegated to the provider
		metric := ocpLogProvider.GetMetric(jsonLog, device.String())
		ch <- metric
	}
}
