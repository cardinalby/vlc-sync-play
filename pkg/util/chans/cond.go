package chans

import (
	"sync"
	"sync/atomic"
)

// Cond is a condition variable implementation with the API similar to as sync.Cond but implemented
// on top of channels.
type Cond struct {
	L        sync.Locker
	isClosed atomic.Bool
	cond     *sync.Cond
}

func NewCond(locker sync.Locker) *Cond {
	return &Cond{
		L:    locker,
		cond: sync.NewCond(locker),
	}
}

// Wait returns a channel that will receive a single value after Signal or Broadcast call.
// After Close() call the channel will be closed without receiving any value.
// You should not call Wait() after Close() call (it will wait forever).
func (c *Cond) Wait() <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		c.cond.Wait()
		if !c.isClosed.Load() {
			ch <- struct{}{}
		}
		close(ch)
	}()
	return ch
}

func (c *Cond) Signal() {
	c.cond.Signal()
}

func (c *Cond) Broadcast() {
	c.cond.Broadcast()
}

// Close closes the condition variable to release internal goroutines.
// Waiting goroutines will not receive any more notifications.
// Signal and Broadcast will not send any more notifications.
func (c *Cond) Close() {
	if !c.isClosed.Swap(true) {
		c.cond.Broadcast()
	}
}
