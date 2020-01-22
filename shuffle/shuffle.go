package shuffle

import (
	"github.com/go-kit/kit/metrics/provider"
	"github.com/hashicorp/golang-lru"
)

type stream struct {
	incoming chan interface{}
	pool     *lru.Cache
	measures *Measures
}

func NewStreamShuffler(poolSize int, bufferSize int, provider provider.Provider) (chan interface{}, func() interface{}) {
	cache, err := lru.New(poolSize)
	if err != nil {
		panic(err)
	}
	shuffler := stream{
		incoming: make(chan interface{}, bufferSize),
		pool:     cache,
		measures: NewMeasures(provider),
	}

	// leverage the incoming channel to fill the pool
	go func() {
		for item := range shuffler.incoming {
			if evicted := shuffler.pool.Add(item, true); !evicted {
				shuffler.measures.DeviceSize.Add(1)
			}
		}
	}()

	return shuffler.incoming, func() interface{} {
		if item, _, ok := shuffler.pool.GetOldest(); !ok {
			return nil
		} else {
			if present := shuffler.pool.Remove(item); present {
				shuffler.measures.DeviceSize.Add(-1)
			}
			return item
		}

	}
}
