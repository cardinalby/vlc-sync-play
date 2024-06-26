package mathutil

import (
	"github.com/gammazero/deque"
)

type AvgAcc[T Numeric] struct {
	maxSamplesCount int
	samples         *deque.Deque[T]
}

func NewAvgAcc[T Numeric](maxSamplesCount int) *AvgAcc[T] {
	return &AvgAcc[T]{
		maxSamplesCount: maxSamplesCount,
		samples:         deque.New[T](maxSamplesCount),
	}
}

func (a *AvgAcc[T]) Add(value T) {
	if a.samples.Len() >= a.maxSamplesCount {
		a.samples.PopFront()
	}
	a.samples.PushBack(value)
}

func (a *AvgAcc[T]) Avg() (T, bool) {
	var sum T
	length := a.samples.Len()
	if length == 0 {
		return 0, false
	}
	for i := 0; i < length; i++ {
		sum += a.samples.At(i)
	}
	return sum / T(length), true
}
