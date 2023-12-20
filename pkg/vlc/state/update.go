package state

import (
	"fmt"

	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic"
)

type Update struct {
	IsNatural    bool
	ChangedProps ChangedProps
	Status       basic.StatusEx
}

func newFullManualUpdate(new *basic.StatusEx) Update {
	upd := Update{
		Status: *new,
	}
	upd.ChangedProps.SetState(true)
	if new.State == basic.PlaybackStateStopped {
		return upd
	}
	upd.ChangedProps.SetFileURI(true)
	upd.ChangedProps.SetPosition(true)
	upd.ChangedProps.SetRate(true)
	upd.IsNatural = false
	return upd
}

func (d *Update) String() string {
	return fmt.Sprintf("natural: %v, %s", d.IsNatural, d.ChangedProps.String())
}
