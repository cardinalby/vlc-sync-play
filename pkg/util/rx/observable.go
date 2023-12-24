package rx

import (
	"sync"
)

type subscription[T any] struct {
	value    *value[T]
	callback func(value T)
}

func (o *subscription[T]) Unsubscribe() {
	o.value.unsubscribe(o)
	o.callback = nil
	o.value = nil
}

type Observers []Subscription

func (o Observers) UnsubscribeAll() {
	for _, observer := range o {
		observer.Unsubscribe()
	}
}

type value[T any] struct {
	mu    sync.RWMutex
	value T
	subs  map[*subscription[T]]struct{}
}

func NewValue[T any](val T) Value[T] {
	return &value[T]{
		value: val,
		subs:  make(map[*subscription[T]]struct{}),
	}
}

func (o *value[T]) GetValue() T {
	return o.value
}

func (o *value[T]) SetValue(value T) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	o.value = value
	for observer := range o.subs {
		observer.callback(value)
	}
}

func (o *value[T]) Subscribe(callback func(value T)) Subscription {
	o.mu.Lock()
	defer o.mu.Unlock()

	observer := &subscription[T]{
		value:    o,
		callback: callback,
	}
	o.subs[observer] = struct{}{}

	return observer
}

func (o *value[T]) unsubscribe(observer *subscription[T]) {
	o.mu.Lock()
	defer o.mu.Unlock()
	delete(o.subs, observer)
}
