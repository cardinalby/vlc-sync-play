package timeutil

import (
	"context"
	"time"
)

// SleepCtx sleeps for the specified duration (returns nil) or until the context is done (returns context error).
func SleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	}

	// using time.After() can lead to resources leak since it can't be cancelled
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
