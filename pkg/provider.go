package pkg

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
)

// MetricProvider is an object that computes the info metric
// from the device data in JSON format.
type MetricProvider struct {
	// Desc holds the pointer to the prometheus.desc object
	Desc *prometheus.Desc

	// ValueType holds the prometheus.ValueType
	ValueType prometheus.ValueType

	// jsonKey is the string key that the object needs to access
	// in the device JSON to fetch the metric float64 value
	jsonKey string

	// valueAsLabel, when true, makes GetMetric emit a constant gauge of 1
	// and attach the string found at jsonKey as the last label instead of
	// parsing it as a float. This is the Prometheus "info metric" pattern,
	// used for identifiers/strings (e.g. a 128-bit GUID) that have no
	// meaningful numeric value.
	valueAsLabel bool
}

// NewMetricProvider is the constructor for MetricProvider objects.
// No need to return a pointer, since the struct is static data.
func NewMetricProvider(
	desc *prometheus.Desc,
	valueType prometheus.ValueType,
	jsonKey string,
) MetricProvider {
	return MetricProvider{
		Desc:      desc,
		ValueType: valueType,
		jsonKey:   jsonKey,
	}
}

// NewLabelMetricProvider builds an "info metric" provider: it emits a constant
// gauge of 1 and attaches the string value at jsonKey as an extra label. Use it
// for identifiers/strings that cannot be represented as a float.
func NewLabelMetricProvider(
	desc *prometheus.Desc,
	jsonKey string,
) MetricProvider {
	return MetricProvider{
		Desc:         desc,
		ValueType:    prometheus.GaugeValue,
		jsonKey:      jsonKey,
		valueAsLabel: true,
	}
}

// GetMetric computes the metric from the
// data in JSON form.
func (ip MetricProvider) GetMetric(
	data gjson.Result,
	labels ...string,
) prometheus.Metric {
	// If data is invalid/empty (e.g., OCP not supported), skip metric creation
	if !data.Exists() {
		return nil
	}

	result := data.Get(ip.jsonKey)

	// If the key is absent (e.g. a field not reported by this firmware/version),
	// skip the metric entirely. Emitting 0 would be misleading and, for counters,
	// would look like a counter reset to Prometheus.
	if !result.Exists() {
		return nil
	}

	// Info-metric mode: emit a constant 1 with the string value as an extra label.
	if ip.valueAsLabel {
		labelValues := make([]string, 0, len(labels)+1)
		labelValues = append(labelValues, labels...)
		labelValues = append(labelValues, result.String())

		return prometheus.MustNewConstMetric(ip.Desc, ip.ValueType, 1, labelValues...)
	}

	// Handle both scalar values (v2.8) and object values (v2.11+)
	// In v2.11+, some fields like critical_warning are objects with a "value" field
	var value float64
	if result.IsObject() {
		// Try to get the "value" field from the object
		value = result.Get("value").Float()
	} else {
		// Direct numeric value
		value = result.Float()
	}

	metric := prometheus.MustNewConstMetric(
		ip.Desc,
		ip.ValueType,
		value,
		labels...,
	)

	return metric
}
