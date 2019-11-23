package shuffle

import (
	"github.com/xmidt-org/capacityset"
)

type stream struct {
	incoming chan interface{}
	pool     capacityset.Set
}

func NewStreamShuffler(poolSize int, bufferSize int) (chan interface{}, func() interface{}) {
	shuffler := stream{
		incoming: make(chan interface{}, bufferSize),
		pool:     capacityset.NewCapacitySet(poolSize),
	}

	// leverage the incoming channel to fill the pool
	go func() {
		for item := range shuffler.incoming {
			shuffler.pool.Add(item)
		}
	}()

	return shuffler.incoming, shuffler.pool.Pop
}
