package shuffle

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetItem(t *testing.T) {
	assert := assert.New(t)

	msg := "Hello, World"

	incoming, getItem := NewStreamShuffler(5, 5)

	incoming <- msg
	time.Sleep(time.Millisecond)
	item := getItem()
	assert.Equal(msg, item)
}

func TestFullPool(t *testing.T) {
	assert := assert.New(t)

	incoming, _ := NewStreamShuffler(5, 2)

	incoming <- 1
	incoming <- 2
	incoming <- 3
	incoming <- 4
	incoming <- 5
	// pool should be full. now and one for the buffer
	incoming <- 6
	// one for transition
	incoming <- 7
	incoming <- 8
	// buffer should now be full.

	select {
	case incoming <- 9:
		assert.Fail("Device Pool should be filled")
	default:
	}

}
