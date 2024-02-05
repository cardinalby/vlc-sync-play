package status_dto

import (
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic"
)

type Status struct {
	LengthSec   int                 `json:"length"`
	Rate        float64             `json:"rate"`
	State       basic.PlaybackState `json:"state"`
	Position    float64             `json:"position"`
	Information struct {
		Category struct {
			Meta struct {
				FileName string `json:"filename"`
			}
		} `json:"category"`
	} `json:"information"`
}

func (s Status) GetFileName() string {
	return s.Information.Category.Meta.FileName
}
