package syncer

import (
	"time"

	"github.com/cardinalby/vlc-sync-play/pkg/util/rx"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/instance"
)

type State struct {
	fileURI                    rx.Value[string]
	lastSyncedAt               time.Time
	acceptFollowerUpdatesAfter time.Time
	lastSyncedFromID           uint
}

func NewState() State {
	return State{
		fileURI:          rx.NewValue(""),
		lastSyncedFromID: instance.IDNone,
	}
}
