package shuffle

import (
	"errors"
	"github.com/Comcast/webpa-common/semaphore"
	mapset "github.com/deckarep/golang-set"
	"sync/atomic"
)

type stream struct {
	incoming chan interface{}
	pool     mapset.Set
	count    *int64
	locking  semaphore.Interface
}

func (s stream) getItem() (interface{}, error) {
	count := atomic.LoadInt64(s.count)
	if count <= 0 {
		return nil, errors.New("no item in pool")
	}
	atomic.AddInt64(s.count, -1)

	// we can accept a new item
	defer func() { s.locking.Release() }()

	return s.pool.Pop(), nil

}

func NewStreamShuffler(poolSize int, bufferSize int) (chan interface{}, func() (interface{}, error)) {
	shuffler := stream{
		incoming: make(chan interface{}, bufferSize),
		pool:     mapset.NewSet(),
		locking:  semaphore.New(poolSize),
		count:    new(int64),
	}

	// leverage the incoming channel to fill the pool
	go func() {
		for {
			shuffler.locking.Acquire()
			item := <-shuffler.incoming
			if !shuffler.pool.Add(item) {
				// item was not added so append to outgoing, so we can try again
				shuffler.locking.Release()
			} else {
				atomic.AddInt64(shuffler.count, 1)
			}
		}

	}()

	return shuffler.incoming, shuffler.getItem
}
