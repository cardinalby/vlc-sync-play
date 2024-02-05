package state

import (
	"time"

	mathutil "github.com/cardinalby/vlc-sync-play/pkg/util/math"
)

// vlcMaxPlaybackStepSeconds is max playback time position change lag that happens on natural playback. Found by experiments
const vlcMaxPlaybackStepSeconds = 0.5

// vlcPositionErrorK is a coefficient for position error calculation. Found by experiments.
// In ideal case it should be 1, but can be more to lower the chance of false positive seek detection.
// Big one (> 10) can lead to false negative seek detection.
// TODO: move to settings
const vlcPositionErrorK = 2

func newPositionRange(statusPosition float64, lengthSec int, rate float64) mathutil.Range[float64] {
	return mathutil.NewRangeMinWithLen(
		statusPosition,
		vlcMaxPlaybackStepSeconds/float64(lengthSec)*max(rate, 1.0)*vlcPositionErrorK,
	)
}

func newPositionDurationRange(positionDuration time.Duration, rate float64) mathutil.Range[time.Duration] {
	return mathutil.NewRangeMinWithLen(
		positionDuration,
		time.Duration(float64(time.Second)*vlcMaxPlaybackStepSeconds*max(rate, 1.0))*vlcPositionErrorK,
	)
}
