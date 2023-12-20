package chans

import (
	"sync"
	"sync/atomic"
)

type BroadcastListener struct {
	c chan struct{}
	b atomic.Pointer[Broadcast]
}

// Unsubscribe unsubscribes from the broadcast. After this call the listener will not receive notifications anymore
// (channel will be closed).
func (l *BroadcastListener) Unsubscribe() {
	if oldBroadcast := l.b.Swap(nil); oldBroadcast != nil {
		oldBroadcast.unsubscribe(l.c)
	}
}

// GetSignalChan returns a channel that will receive a single value after Broadcast.Notify() call.
// After Unsubscribe() call the channel will be closed.
func (l *BroadcastListener) GetSignalChan() <-chan struct{} {
	return l.c
}

// Broadcast allows to notify multiple goroutines about some event
type Broadcast struct {
	listeners map[chan struct{}]struct{}
	mu        sync.RWMutex
}

func NewBroadcast() *Broadcast {
	return &Broadcast{
		listeners: make(map[chan struct{}]struct{}),
	}
}

// Subscribe returns a listener that can be used to wait for the next notification. The caller should
// call Unsubscribe() on listener when it doesn't need to receive notifications anymore.
func (b *Broadcast) Subscribe() *BroadcastListener {
	c := make(chan struct{}, 1)

	b.mu.Lock()
	b.listeners[c] = struct{}{}
	b.mu.Unlock()

	res := BroadcastListener{c: c}
	res.b.Store(b)
	return &res
}

// Notify asynchronously notifies all listeners about the event
func (b *Broadcast) Notify() {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for c := range b.listeners {
		select {
		case c <- struct{}{}:
		default:
		}
	}
}

func (b *Broadcast) unsubscribe(c chan struct{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.listeners, c)
	close(c)
}
