package main

import (
	"github.com/Comcast/webpa-common/xmetrics"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/provider"
)

const (
	CompletedCounter = "completed_count"
	SuccessCounter   = "success_count"
	DeviceSetSize    = "device_set_size"
)

func Metrics() []xmetrics.Metric {
	return []xmetrics.Metric{
		{
			Name: CompletedCounter,
			Help: "The total number of completed request to both xmidt and gungnir",
			Type: "counter",
		},
		{
			Name: SuccessCounter,
			Help: "The total number of success",
			Type: "counter",
		},
		{
			Name: DeviceSetSize,
			Help: "The total number of devices in the set",
			Type: "gauge",
		},
	}
}

type Measures struct {
	Completed  metrics.Counter
	Success    metrics.Counter
	DeviceSize metrics.Gauge
}

// NewMeasures constructs a Measures given a go-kit metrics Provider
func NewMeasures(p provider.Provider) *Measures {
	return &Measures{
		Completed:  p.NewCounter(CompletedCounter),
		Success:    p.NewCounter(SuccessCounter),
		DeviceSize: p.NewGauge(DeviceSetSize),
	}
}
