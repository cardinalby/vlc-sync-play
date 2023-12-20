package extended

import (
	"time"

	typeutil "github.com/cardinalby/vlc-sync-play/pkg/util/type"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic"
)

type ExpectedPositionGetter func(atMoment time.Time) float64

type CmdGroup struct {
	OpenFile typeutil.Optional[string]
	Seek     typeutil.Optional[ExpectedPositionGetter]
	Rate     typeutil.Optional[float64]
	State    typeutil.Optional[basic.PlaybackState]
}

func (g CmdGroup) GetOpenFileCmd() basic.Command {
	if !g.OpenFile.HasValue {
		return nil
	}
	return basic.PlayFileCmd(g.OpenFile.Value)
}

func (g CmdGroup) GetSeekCmd(expectedExecutionTime time.Time) basic.Command {
	if !g.Seek.HasValue {
		return nil
	}
	return basic.SeekCmd(g.Seek.Value(expectedExecutionTime))
}

func (g CmdGroup) GetRateCmd() basic.Command {
	if !g.Rate.HasValue {
		return nil
	}
	return basic.RateCmd(g.Rate.Value)
}

func (g CmdGroup) GetStateCmd() basic.Command {
	if !g.State.HasValue {
		return nil
	}
	switch g.State.Value {
	case basic.PlaybackStatePlaying:
		return basic.ResumeCmd()
	case basic.PlaybackStatePaused:
		return basic.PauseCmd()
	case basic.PlaybackStateStopped:
		return basic.StopCmd()
	}
	return nil
}

func (g CmdGroup) HasAny() bool {
	return g.OpenFile.HasValue || g.Seek.HasValue || g.Rate.HasValue || g.State.HasValue
}
