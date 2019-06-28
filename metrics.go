package main

import (
	"github.com/Comcast/webpa-common/xmetrics"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/provider"
)

const (
	CompletedCounter = "completed_count"
	SuccessCounter   = "success_count"
	FailureCounter   = "failure_count"
	DeviceSetSize    = "device_set_size"
)

const (
	// StateLabel is for labeling metrics; it provides the device status
	// according to xmidt.
	StateLabel   = "connected_state"
	OnlineState  = "online"
	OfflineState = "offline"
)

func Metrics() []xmetrics.Metric {
	return []xmetrics.Metric{
		{
			Name: CompletedCounter,
			Help: "The total number of completed request to both xmidt and codex",
			Type: "counter",
		},
		{
			Name:       SuccessCounter,
			Help:       "The total number of success",
			Type:       "counter",
			LabelNames: []string{StateLabel},
		},
		{
			Name:       FailureCounter,
			Help:       "The total number of failure",
			Type:       "counter",
			LabelNames: []string{StateLabel},
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
	Failure    metrics.Counter
	DeviceSize metrics.Gauge
}

// NewMeasures constructs a Measures given a go-kit metrics Provider
func NewMeasures(p provider.Provider) *Measures {
	return &Measures{
		Completed:  p.NewCounter(CompletedCounter),
		Success:    p.NewCounter(SuccessCounter),
		Failure:    p.NewCounter(FailureCounter),
		DeviceSize: p.NewGauge(DeviceSetSize),
	}
}
