package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetrics(t *testing.T) {
	assert := assert.New(t)

	m := Metrics()

	assert.NotNil(m)
}
