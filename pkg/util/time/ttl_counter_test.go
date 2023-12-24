package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewTtlCounterSingleInc(t *testing.T) {
	t.Parallel()

	c := NewTtlCounter(time.Millisecond * 10)
	c.Inc()
	time.Sleep(time.Millisecond * 5)
	require.True(t, c.TryDec())
	require.Equal(t, int64(0), c.counter)
	require.False(t, c.TryDec())
	require.Equal(t, int64(0), c.counter)
}

func TestNewTtlCounterExpiredInc(t *testing.T) {
	t.Parallel()

	c := NewTtlCounter(time.Millisecond * 10)
	c.Inc()
	time.Sleep(time.Millisecond * 15)
	require.False(t, c.TryDec())
	require.Equal(t, int64(0), c.counter)
}

func TestNewTtlCounterMultipleSimultaneousInc(t *testing.T) {
	t.Parallel()

	c := NewTtlCounter(time.Millisecond * 10)
	c.Inc()
	c.Inc()
	time.Sleep(time.Millisecond * 5)
	require.Equal(t, int64(2), c.counter)
	require.True(t, c.TryDec())
	require.Equal(t, int64(1), c.counter)
	require.True(t, c.TryDec())
	require.Equal(t, int64(0), c.counter)
	require.False(t, c.TryDec())
}

func TestNewTtlCounterMultipleDelayedInc(t *testing.T) {
	t.Parallel()

	c := NewTtlCounter(time.Millisecond * 10)
	c.Inc()
	time.Sleep(time.Millisecond * 5)
	c.Inc()
	require.Equal(t, int64(2), c.counter)

	require.True(t, c.TryDec())
	require.Equal(t, int64(1), c.counter)
	time.Sleep(time.Millisecond * 8)
	require.True(t, c.TryDec())
	require.Equal(t, int64(0), c.counter)
}

func TestNewTtlCounterIncWhenExpired(t *testing.T) {
	t.Parallel()

	c := NewTtlCounter(time.Millisecond * 10)
	c.Inc()
	time.Sleep(time.Millisecond * 15)
	c.Inc()
	require.Equal(t, int64(1), c.counter)
	time.Sleep(time.Millisecond * 5)
	require.True(t, c.TryDec())
	require.Equal(t, int64(0), c.counter)
	require.False(t, c.TryDec())
	require.Equal(t, int64(0), c.counter)
}
