package chans

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWaitSignalsWhenConditionIsMet(t *testing.T) {
	locker := &sync.Mutex{}
	cond := NewCond(locker)

	go func() {
		time.Sleep(100 * time.Millisecond)
		locker.Lock()
		defer locker.Unlock()
		cond.Signal()
	}()

	locker.Lock()
	select {
	case <-cond.Wait():
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Wait did not return after Signal was called")
	}
}

func TestSignalWakesUpOneWaitingGoroutine(t *testing.T) {
	t.Parallel()

	locker := &sync.Mutex{}
	cond := NewCond(locker)

	var res []int

	for i := 0; i < 10; i++ {
		i := i
		go func() {
			locker.Lock()
			_, ok := <-cond.Wait()
			if ok {
				res = append(res, i)
			}
			locker.Unlock()
		}()
	}

	time.Sleep(100 * time.Millisecond)

	locker.Lock()
	cond.Signal()
	locker.Unlock()

	time.Sleep(100 * time.Millisecond)

	assert.Len(t, res, 1)

	locker.Lock()
	cond.Signal()
	locker.Unlock()

	time.Sleep(100 * time.Millisecond)

	assert.Len(t, res, 2)

	cond.Close()

	time.Sleep(100 * time.Millisecond)

	assert.Len(t, res, 2)
}

func TestBroadcastWakesUpAllWaitingGoroutines(t *testing.T) {
	t.Parallel()

	locker := &sync.Mutex{}
	cond := NewCond(locker)

	var res []int

	for i := 0; i < 10; i++ {
		i := i
		go func() {
			locker.Lock()
			_, ok := <-cond.Wait()
			if ok {
				res = append(res, i)
			}
			locker.Unlock()
		}()
	}

	time.Sleep(100 * time.Millisecond)

	locker.Lock()
	cond.Broadcast()
	locker.Unlock()

	time.Sleep(100 * time.Millisecond)

	assert.Len(t, res, 10)
}

func TestClosePreventsFurtherSignals(t *testing.T) {
	t.Parallel()

	locker := &sync.Mutex{}
	cond := NewCond(locker)

	var res []int
	doneGoroutines := atomic.Int32{}

	for i := 0; i < 10; i++ {
		i := i
		go func() {
			locker.Lock()
			defer locker.Unlock()
			_, ok := <-cond.Wait()
			if ok {
				res = append(res, i)
			}
			doneGoroutines.Add(1)
		}()
	}

	time.Sleep(100 * time.Millisecond)

	cond.Close()
	locker.Lock()
	cond.Broadcast()
	locker.Unlock()

	time.Sleep(100 * time.Millisecond)

	assert.Len(t, res, 0)
	assert.Equal(t, int32(10), doneGoroutines.Load())
}
