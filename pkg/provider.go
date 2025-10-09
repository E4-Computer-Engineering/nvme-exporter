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
