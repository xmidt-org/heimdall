package shuffle

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/xmidt-org/webpa-common/v2/xmetrics/xmetricstest"
)

func TestGetItem(t *testing.T) {
	assert := assert.New(t)

	msg := "Hello, World"
	metricsProvider := xmetricstest.NewProvider(nil, Metrics)
	steam := NewStreamShuffler(5, metricsProvider)

	steam.Add(msg)
	metricsProvider.Assert(t, DeviceSetSize)(xmetricstest.Value(1))
	time.Sleep(time.Millisecond)
	item := steam.Get()
	assert.Equal(msg, item)
	metricsProvider.Assert(t, DeviceSetSize)(xmetricstest.Value(0))
}

func TestFullPool(t *testing.T) {
	assert := assert.New(t)
	metricsProvider := xmetricstest.NewProvider(nil, Metrics)
	steam := NewStreamShuffler(5, metricsProvider)

	for i := 0; i < 10; i++ {
		go steam.Add(i)
	}
	time.Sleep(time.Millisecond)

	metricsProvider.Assert(t, DeviceSetSize)(xmetricstest.Value(5))
	data := map[interface{}]int{}
	item := steam.Get()
	assert.NotNil(item)
	data[item]++

	time.Sleep(time.Millisecond)
	metricsProvider.Assert(t, DeviceSetSize)(xmetricstest.Value(5))
	for i := 0; i < 5; i++ {
		item := steam.Get()
		assert.NotNil(item)
		data[item]++
	}
	time.Sleep(time.Millisecond)

	metricsProvider.Assert(t, DeviceSetSize)(xmetricstest.Value(4))
	for i := 0; i < 4; i++ {
		item := steam.Get()
		assert.NotNil(item)
		data[item]++
	}
	metricsProvider.Assert(t, DeviceSetSize)(xmetricstest.Value(0))
	for key, value := range data {
		assert.Equal(1, value, fmt.Sprintf("error with key %s", key))
	}
}
