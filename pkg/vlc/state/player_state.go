package state

import (
	"errors"
	"sync"
	"time"

	mathutil "github.com/cardinalby/vlc-sync-play/pkg/util/math"
	timeutil "github.com/cardinalby/vlc-sync-play/pkg/util/time"
	typeutil "github.com/cardinalby/vlc-sync-play/pkg/util/type"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/extended"
)

type playbackBase struct {
	moment   timeutil.Range
	pbTimeR  mathutil.Range[time.Duration]
	position float64
	rate     float64
}

type State struct {
	mu     sync.RWMutex
	pbBase typeutil.Optional[playbackBase]
	prev   typeutil.Optional[basic.StatusEx]
}

var errOlderThenPrevious = errors.New("new status is older than prev status")

func NewState() *State {
	return &State{}
}

func (s *State) ApplyNewStatus(new *basic.StatusEx) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.applyNewStatus(new)
}

func (s *State) ApplyNewStatusAndGetUpdate(new *basic.StatusEx) (Update, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer s.applyNewStatus(new)
	return s.getUpdate(new)
}

func (s *State) GetUpdate(new *basic.StatusEx) (Update, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.getUpdate(new)
}

func (s *State) GetSyncCommands(props ChangedProps) extended.CmdGroup {
	cmdGr := extended.CmdGroup{}
	if !props.HasAny() {
		return cmdGr
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	prev := &s.prev.Value

	if props.HasFileURI() {
		cmdGr.OpenFile.Set(prev.FileURI)
	}

	if props.HasPosition() ||
		(props.HasState() && prev.State != basic.PlaybackStateStopped) {
		if expectedPositionGetter := s.GetExpectedPosition(); expectedPositionGetter != nil {
			cmdGr.Seek.Set(expectedPositionGetter)
		}
	}
	if props.HasRate() {
		cmdGr.Rate.Set(prev.Rate)
	}
	if props.HasState() {
		cmdGr.State.Set(prev.State)
	}

	return cmdGr
}

func (s *State) GetExpectedPosition() extended.ExpectedPositionGetter {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.prev.HasValue || s.prev.Value.State == basic.PlaybackStateStopped {
		return nil
	}
	prev := &s.prev.Value
	if prev.State == basic.PlaybackStatePaused {
		result := newPositionRange(prev.Position, prev.LengthSec).Center()
		return func(_ time.Time) float64 {
			return result
		}
	}
	prevPositionRange := func() mathutil.Range[float64] {
		if pbBase := &s.pbBase.Value; s.pbBase.HasValue && pbBase.position != prev.Position {
			// can narrow down the range
			timeSincePbBase := prev.Moment.SubRange(pbBase.moment)
			expectedPrevPbTime := pbBase.pbTimeR.AddRange(timeSincePbBase.MultiplyF(pbBase.rate))
			actualPrevPbTime := newPositionDurationRange(prev.GetPbTime())
			pbTimeIntersection, ok := expectedPrevPbTime.Intersection(actualPrevPbTime)
			if ok {
				// should be always true
				return pbTimeIntersection.DivF(float64(prev.GetLength()))
			}
		}
		return newPositionRange(prev.Position, prev.LengthSec)
	}()

	prevMoment := prev.Moment
	durationToPosMultiplier := prev.Rate / float64(prev.GetLength())

	return func(atMoment time.Time) float64 {
		timeSincePrev := timeutil.NewRangeWithLen(atMoment, 0).SubRange(prevMoment)

		return prevPositionRange.AddRange(
			timeSincePrev.ToFloat64().MultiplyF(durationToPosMultiplier),
		).Center()
	}
}

func (s *State) applyNewStatus(new *basic.StatusEx) {
	s.prev.Set(*new)

	isPlaying := new.State == basic.PlaybackStatePlaying
	if s.pbBase.HasValue && !isPlaying {
		s.pbBase.Reset()
	} else if isPlaying && (!s.pbBase.HasValue ||
		s.pbBase.Value.position != new.Position ||
		s.pbBase.Value.rate != new.Rate) {
		// update pbBase
		s.pbBase.Set(playbackBase{
			moment:   new.Moment,
			position: new.Position,
			rate:     new.Rate,
			pbTimeR:  newPositionDurationRange(new.GetPbTime()),
		})
	}
}

func (s *State) getUpdate(new *basic.StatusEx) (Update, error) {
	if s.prev.HasValue {
		return getUpdateFromPrev(&s.prev.Value, &s.pbBase, new)
	}
	return newFullManualUpdate(new), nil
}

func getUpdateFromPrev(
	prev *basic.StatusEx,
	pbBase *typeutil.Optional[playbackBase],
	new *basic.StatusEx,
) (Update, error) {
	upd := Update{
		Status:    *new,
		IsNatural: false,
	}

	if new.Moment.Min.Before(prev.Moment.Max) {
		return upd, errOlderThenPrevious
	}

	upd.ChangedProps.SetFileURI(prev.FileURI != new.FileURI)
	upd.ChangedProps.SetState(new.State != prev.State)
	upd.ChangedProps.SetRate(prev.Rate != new.Rate)

	naturalPositionChange := false
	if new.Position != prev.Position {
		upd.ChangedProps.SetPosition(true)
		if new.State == basic.PlaybackStatePlaying &&
			pbBase.HasValue &&
			new.Position >= prev.Position {
			// can be a natural playback
			actualPbTimeDelta := getActualPlaybackTimeDeltaFromPbBase(&pbBase.Value, new)
			expectedPbTimeDelta := getExpectedPlaybackTimeDeltaFromPbBase(&pbBase.Value, prev, new)
			if expectedPbTimeDelta.HasIntersection(actualPbTimeDelta) {
				naturalPositionChange = true
			}
		}
	}
	upd.IsNatural = !upd.ChangedProps.HasAny() || naturalPositionChange

	return upd, nil
}

func getActualPlaybackTimeDeltaFromPbBase(
	pbBase *playbackBase,
	new *basic.StatusEx,
) mathutil.Range[time.Duration] {
	if pbBase.position == new.Position {
		return newPositionDurationRange(0).MultiplyF(new.Rate)
	}
	newTimeR := newPositionDurationRange(new.GetPbTime())

	return newTimeR.SubRange(pbBase.pbTimeR)
}

func getExpectedPlaybackTimeDeltaFromPbBase(
	pbBase *playbackBase,
	prev *basic.StatusEx,
	new *basic.StatusEx,
) mathutil.Range[time.Duration] {
	if new.Rate == prev.Rate {
		return new.Moment.
			SubRange(pbBase.moment).
			MultiplyF(new.Rate)
	}

	prevToNew := new.Moment.
		SubRange(prev.Moment).
		MultiplyRangeF(mathutil.NewRangeUnordered(new.Rate, prev.Rate))

	if !pbBase.moment.Equal(prev.Moment) {
		baseToPrev := prev.Moment.
			SubRange(pbBase.moment).
			MultiplyF(pbBase.rate)
		return baseToPrev.AddRange(prevToNew)
	}

	return prevToNew
}
