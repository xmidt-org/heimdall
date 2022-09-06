package shuffle

import (
	mapset "github.com/deckarep/golang-set"
	"github.com/go-kit/kit/metrics/provider"
	"github.com/xmidt-org/webpa-common/v2/semaphore"
)

type Interface interface {
	Add(interface{})
	Get() interface{}
}

type stream struct {
	set       mapset.Set
	measures  *Measures
	semaphore semaphore.Interface
}

func (s *stream) Add(key interface{}) {
	if s.set.Contains(key) {
		return
	}
	s.semaphore.Acquire()
	if ok := s.set.Add(key); ok {
		s.measures.DeviceSize.Add(1.0)
	}
}

func (s *stream) Get() interface{} {
	var value interface{}
	// keep popping til value is not nil
	for value == nil {
		value = s.set.Pop()
	}
	s.semaphore.Release()
	s.measures.DeviceSize.Add(-1.0)
	return value
}

func NewStreamShuffler(poolSize int, provider provider.Provider) Interface {

	shuffler := stream{
		semaphore: semaphore.New(poolSize),
		set:       mapset.NewSet(),
		measures:  NewMeasures(provider),
	}

	return &shuffler
}
