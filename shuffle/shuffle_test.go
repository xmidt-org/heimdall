package shuffle

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/xmidt-org/webpa-common/xmetrics/xmetricstest"
	"testing"
	"time"
)

func TestGetItem(t *testing.T) {
	assert := assert.New(t)

	msg := "Hello, World"

	incoming, getItem := NewStreamShuffler(5, 5, xmetricstest.NewProvider(nil, nil))

	incoming <- msg
	time.Sleep(time.Millisecond)
	item := getItem()
	assert.Equal(msg, item)
}

func TestFullPool(t *testing.T) {
	//assert := assert.New(t)

	incoming, pop := NewStreamShuffler(5, 1, xmetricstest.NewProvider(nil, nil))
	a := func() {
		for {
			fmt.Println("a", pop())
		}
	}

	b := func() {
		for {
			fmt.Println("b", pop())
		}
	}
	go a()
	go b()


	incoming <- 1
	incoming <- 2
	incoming <- 3
	incoming <- 4
	incoming <- 1
	// pool should be full. now and one for the buffer
	incoming <- 1
	// one for transition
	incoming <- 1
	incoming <- 2
	// buffer should now be full.
	incoming <- 1
	incoming <- 2
	incoming <- 1
	incoming <- 2
	incoming <- 1
	incoming <- 2
	incoming <- 1
	incoming <- 2
	incoming <- 1
	incoming <- 2
	incoming <- 1
	incoming <- 2
	//select {
	//case incoming <- 9:
	//	assert.Fail("Device Pool should be filled")
	//default:
	//}


	time.Sleep(time.Second *10)
}
