package state

import (
	"time"

	mathutil "github.com/cardinalby/vlc-sync-play/pkg/util/math"
)

const vlcMaxPlaybackStepSeconds = 0.5

// VlcMaxPlaybackStepDuration is max playback time position change that happens on natural playback. Found by experiments
const VlcMaxPlaybackStepDuration = time.Duration(float64(time.Second) * vlcMaxPlaybackStepSeconds)

func newPositionRange(statusPosition float64, lengthSec int) mathutil.Range[float64] {
	return mathutil.NewRangeMinWithLen(
		statusPosition,
		vlcMaxPlaybackStepSeconds/float64(lengthSec),
	)
}

func newPositionDurationRange(positionDuration time.Duration) mathutil.Range[time.Duration] {
	return mathutil.NewRangeMinWithLen(
		positionDuration,
		VlcMaxPlaybackStepDuration,
	)
}
