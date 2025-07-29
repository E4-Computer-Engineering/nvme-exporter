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
	value := data.Get(ip.jsonKey).Float()

	metric := prometheus.MustNewConstMetric(
		ip.Desc,
		ip.ValueType,
		value,
		labels...,
	)

	return metric
}
