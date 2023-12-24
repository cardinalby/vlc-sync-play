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

func (d *Update) String() string {
	if !d.IsNatural {
		return d.ChangedProps.String()
	}
	return fmt.Sprintf("NATURAL: %s", d.ChangedProps.String())
}
