package httpjson

import (
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic"
)

type statusDto struct {
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

func (s statusDto) getFileName() string {
	return s.Information.Category.Meta.FileName
}

type playlistItemType string

const playlistItemTypeNode playlistItemType = "node"
const playlistItemTypeLeaf playlistItemType = "leaf"

type playlistItemDto struct {
	Type     playlistItemType  `json:"type"`
	Name     string            `json:"name"`
	Uri      string            `json:"uri"`
	Current  string            `json:"current"`
	Children []playlistItemDto `json:"children"`
}

func (pl playlistItemDto) getCurrent() (playlistItemDto, bool) {
	if pl.Type == playlistItemTypeLeaf {
		return pl, pl.Current != ""
	}
	for _, child := range pl.Children {
		if curr, ok := child.getCurrent(); ok {
			return curr, true
		}
	}

	return pl, false
}
