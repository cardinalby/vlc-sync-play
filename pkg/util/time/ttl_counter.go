package timeutil

import (
	"sync"
	"time"
)

type TtlCounter struct {
	mu          sync.Mutex
	counter     int64
	lastAddedAt time.Time
	ttl         time.Duration
}

func NewTtlCounter(ttl time.Duration) *TtlCounter {
	return &TtlCounter{
		ttl: ttl,
	}
}

func (t *TtlCounter) Inc() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.lastAddedAt.Add(t.ttl).Before(time.Now()) {
		t.counter = 0
	}
	t.counter++
	t.lastAddedAt = time.Now()
}

func (t *TtlCounter) TryDec() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.counter == 0 {
		return false
	}
	if t.lastAddedAt.Add(t.ttl).Before(time.Now()) {
		t.counter = 0
		return false
	}
	t.counter--
	return true
}
