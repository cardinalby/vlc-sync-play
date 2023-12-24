package state

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cardinalby/vlc-sync-play/pkg/util/logging"
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
	mu             sync.RWMutex
	fileJustOpened bool
	pbBase         typeutil.Optional[playbackBase]
	prev           typeutil.Optional[basic.StatusEx]
	logger         logging.Logger
}

var errOlderThenPrevious = errors.New("new status is older than prev status")

func NewState(logger logging.Logger) *State {
	return &State{
		logger: logger,
	}
}

func (s *State) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sb := &strings.Builder{}
	if s.fileJustOpened {
		sb.WriteString("fileJustOpened ")
	}
	if s.pbBase.HasValue {
		sb.WriteString(fmt.Sprintf("pbBase[pos: %v, rate: %v] ", s.pbBase.Value.position, s.pbBase.Value.rate))
	}
	if s.prev.HasValue {
		sb.WriteString(fmt.Sprintf("prev[%s]", s.prev.Value.String()))
	}
	return sb.String()
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

func (s *State) GetPauseOrResumeCommand() extended.CmdGroup {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cmdGr := extended.CmdGroup{}
	if !s.prev.HasValue {
		return cmdGr
	}
	prev := &s.prev.Value

	switch prev.State {
	case basic.PlaybackStatePlaying:
		cmdGr.State.Set(basic.PlaybackStatePaused)
	case basic.PlaybackStatePaused:
		cmdGr.State.Set(basic.PlaybackStatePlaying)
	default:
		return cmdGr
	}

	if expectedPositionGetter := s.GetExpectedPosition(); expectedPositionGetter != nil {
		cmdGr.Seek.Set(expectedPositionGetter)
	}
	return cmdGr
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
		result := newPositionRange(prev.Position, prev.LengthSec, prev.Rate).Center()
		return func(_ time.Time) float64 {
			return result
		}
	}
	prevPositionRange := func() mathutil.Range[float64] {
		if pbBase := &s.pbBase.Value; s.pbBase.HasValue && pbBase.position != prev.Position {
			// can narrow down the range
			timeSincePbBase := prev.Moment.SubRange(pbBase.moment)
			expectedPrevPbTime := pbBase.pbTimeR.AddRange(timeSincePbBase.MultiplyF(pbBase.rate))
			actualPrevPbTime := newPositionDurationRange(prev.GetPbTime(), prev.Rate)
			pbTimeIntersection, ok := expectedPrevPbTime.Intersection(actualPrevPbTime)
			if ok {
				// should be always true
				return pbTimeIntersection.DivF(float64(prev.GetLength()))
			}
		}
		return newPositionRange(prev.Position, prev.LengthSec, prev.Rate)
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
	if s.prev.HasValue && new.Moment.Min.Before(s.prev.Value.Moment.Max) {
		return // old status
	}

	if (!s.prev.HasValue && new.FileURI != "") ||
		(s.prev.HasValue && s.prev.Value.FileURI != new.FileURI) {
		// a new file opened
		s.prev.Set(*new)
		s.fileJustOpened = true
		s.pbBase.Reset()
		return
	}

	s.fileJustOpened = false
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
			pbTimeR:  newPositionDurationRange(new.GetPbTime(), new.Rate),
		})
	}
}

func (s *State) getUpdate(new *basic.StatusEx) (Update, error) {
	if s.prev.HasValue {
		return s.getUpdateFromPrev(new)
	}
	return s.getInitStatusUpdate(new), nil
}

func (s *State) getUpdateFromPrev(new *basic.StatusEx) (Update, error) {
	prev := &s.prev.Value
	pbBase := &s.pbBase

	upd := Update{
		Status:    *new,
		IsNatural: false,
	}

	if new.Moment.Min.Before(prev.Moment.Max) {
		return upd, errOlderThenPrevious
	}

	if new.FileURI != prev.FileURI {
		upd.ChangedProps.SetFileURI(true)
		return upd, nil
	}

	if s.fileJustOpened {
		upd.ChangedProps.SetPosition(true)
		upd.ChangedProps.SetRate(true)
		upd.ChangedProps.SetState(true)
		return upd, nil
	}

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
			} else {
				s.logger.Info(
					"NOT NATURAL: expected pb time delta: %s, actual: %s",
					expectedPbTimeDelta, actualPbTimeDelta)
			}
		}
	}
	upd.IsNatural = !upd.ChangedProps.HasRate() &&
		!upd.ChangedProps.HasState() &&
		!upd.ChangedProps.HasFileURI() &&
		(!upd.ChangedProps.HasPosition() || naturalPositionChange)

	return upd, nil
}

func (s *State) getInitStatusUpdate(new *basic.StatusEx) Update {
	upd := Update{
		Status: *new,
	}
	if new.State == basic.PlaybackStateStopped || new.FileURI == "" {
		return upd
	}
	upd.ChangedProps.SetFileURI(true)
	return upd
}

func getActualPlaybackTimeDeltaFromPbBase(
	pbBase *playbackBase,
	new *basic.StatusEx,
) mathutil.Range[time.Duration] {
	if pbBase.position == new.Position {
		return newPositionDurationRange(0, new.Rate)
	}
	newTimeR := newPositionDurationRange(new.GetPbTime(), new.Rate)

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
