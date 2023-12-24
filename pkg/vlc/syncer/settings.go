package syncer

import (
	"time"

	"github.com/cardinalby/vlc-sync-play/pkg/util/rx"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/instance"
)

type Settings interface {
	GetInstancesNumber() rx.Observable[int]
	GetNoVideo() rx.Observable[bool]
	GetPollingInterval() rx.Observable[time.Duration]
	GetClickPause() rx.Observable[bool]
	GetReSeekSrc() rx.Observable[bool]
}

func getPlayerSettings(s Settings) playerSettings {
	return playerSettings{
		pollingInterval: s.GetPollingInterval(),
		stdErrEvents: rx.Map(s.GetClickPause(), func(value bool) instance.EventsToParse {
			if value {
				return instance.EventsToParse{instance.StderrEventMouse1Click: true}
			} else {
				return instance.EventsToParse{}
			}
		}),
	}
}
