// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package shuffle

import (
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/provider"
	"github.com/xmidt-org/webpa-common/v2/xmetrics" // nolint: staticcheck
)

const (
	DeviceSetSize = "device_set_size"
)

func Metrics() []xmetrics.Metric {
	return []xmetrics.Metric{
		{
			Name: DeviceSetSize,
			Help: "The total number of devices in the set",
			Type: "gauge",
		},
	}
}

type Measures struct {
	DeviceSize metrics.Gauge
}

// NewMeasures constructs a Measures given a go-kit metrics Provider
func NewMeasures(p provider.Provider) *Measures {
	return &Measures{
		DeviceSize: p.NewGauge(DeviceSetSize),
	}
}
