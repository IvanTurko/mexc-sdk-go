package sync

import "sync/atomic"

type Counter interface {
	Get() uint64
	Inc() uint64
}

type CounterImpl struct {
	value uint64
}

func NewCounter() Counter {
	return &CounterImpl{}
}

func (c *CounterImpl) Get() uint64 {
	return atomic.LoadUint64(&c.value)
}

func (c *CounterImpl) Inc() uint64 {
	return atomic.AddUint64(&c.value, 1)
}
