package basic

import (
	"fmt"
	"time"

	timeutil "github.com/cardinalby/vlc-sync-play/pkg/util/time"
)

type PlaybackState string

const PlaybackStatePlaying PlaybackState = "playing"
const PlaybackStatePaused PlaybackState = "paused"
const PlaybackStateStopped PlaybackState = "stopped"

type Status struct {
	LengthSec int
	Rate      float64
	State     PlaybackState
	Position  float64
	FileName  string
	Moment    timeutil.Range
}

func (s Status) GetPbTime() time.Duration {
	return time.Duration(s.Position * float64(s.LengthSec) * float64(time.Second))
}

func (s Status) GetLength() time.Duration {
	return time.Duration(s.LengthSec) * time.Second
}

type StatusEx struct {
	Status
	FileURI string
}

func (s StatusEx) String() string {
	return fmt.Sprintf("Status{LengthSec: %d, Rate: %f, State: %s, Position: %f, FileURI: %s, Moment: %s - %s}",
		s.LengthSec, s.Rate, s.State, s.Position, s.FileURI, s.Moment.Min, s.Moment.Max)
}
