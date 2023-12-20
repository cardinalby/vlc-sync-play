package timeutil

import (
	"time"

	mathutil "github.com/cardinalby/vlc-sync-play/pkg/util/math"
)

type Range struct {
	Min time.Time
	Max time.Time
}

func (r Range) SubRange(another Range) mathutil.Range[time.Duration] {
	return mathutil.Range[time.Duration]{
		Min: r.Min.Sub(another.Max),
		Max: r.Max.Sub(another.Min),
	}
}

func (r Range) Center() time.Time {
	return r.Min.Add(r.Length() / 2)
}

func (r Range) Length() time.Duration {
	return r.Max.Sub(r.Min)
}

func (r Range) Equal(other Range) bool {
	return r.Min.Equal(other.Min) && r.Max.Equal(other.Max)
}

func NewRangeWithLen(min time.Time, len time.Duration) Range {
	return Range{
		Min: min,
		Max: min.Add(len),
	}
}
