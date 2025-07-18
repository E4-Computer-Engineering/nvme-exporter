package pkg

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
)

// DescData is a simple struct that pairs a pointer to a
// prometheus.Desc and a prometheus.ValueType.
type DescData struct {

	// desc holds the pointer to the prometheus.desc object
	Desc *prometheus.Desc

	// ValueType holds the prometheus.ValueType
	ValueType prometheus.ValueType
}

// InfoMetricProvider is an object that computes the info metric
// from the device data in JSON format
type InfoMetricProvider struct {

	// DescData is embedded
	DescData

	// JsonKey is the string key that the object needs to access
	// in the device JSON to fetch the metric float64 value
	JsonKey string
}

// NewInfoMetricProvider is the constructor for InfoMetricProvider objects.
// It directly initializes the DescData struct and returns
// the built InfoMetricProvider.
// No need to return a pointer, since the struct is mostly static data.
func NewInfoMetricProvider(
	desc *prometheus.Desc,
	valueType prometheus.ValueType,
	jsonKey string,
) InfoMetricProvider {
	return InfoMetricProvider{
		DescData: DescData{
			Desc:      desc,
			ValueType: valueType,
		},
		JsonKey: jsonKey,
	}
}

// GetMetric computes the metric from the
// device data in JSON form
func (ip InfoMetricProvider) GetMetric(
	device gjson.Result,
	labels ...string,
) prometheus.Metric {
	value := device.Get(ip.JsonKey).Float()

	metric := prometheus.MustNewConstMetric(
		ip.Desc,
		ip.ValueType,
		value,
		labels...,
	)

	return metric
}

// LogMetricProvider is an object that computes the metric
// from the smart log data in JSON format
type LogMetricProvider struct {
	// DescData is embedded
	DescData

	// JsonKey is the string key that the object needs to access
	// in the logs JSON to fetch the metric float64 value
	JsonKey string
}

// NewLogMetricProvider is the constructor for LogMetricProvider objects.
// It directly initializes the DescData struct and returns
// the built LogMetricProvider.
// No need to return a pointer, since the struct is mostly static data.
func NewLogMetricProvider(
	desc *prometheus.Desc,
	valueType prometheus.ValueType,
	jsonKey string,
) LogMetricProvider {
	return LogMetricProvider{
		DescData: DescData{
			Desc:      desc,
			ValueType: valueType,
		},
		JsonKey: jsonKey,
	}
}

// GetMetric computes the metric from the
// smart log data in JSON form
func (lp LogMetricProvider) GetMetric(
	smartLog gjson.Result,
	labels ...string,
) prometheus.Metric {
	value := smartLog.Get(lp.JsonKey).Float()

	metric := prometheus.MustNewConstMetric(
		lp.Desc,
		lp.ValueType,
		value,
		labels...,
	)

	return metric
}
