package playlist_dto

import (
	"encoding/json"
	"errors"
)

// VersionsUnmarshaller can unmarshal both single item and array of items (probably the request format depends
// on the VLC version) to one ItemDto
type VersionsUnmarshaller struct {
	ItemDto ItemDto
}

func (u *VersionsUnmarshaller) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return errors.New("empty data")
	}

	// array of items
	if data[0] == '[' {
		u.ItemDto.Type = ItemTypeNode
		return json.Unmarshal(data, &u.ItemDto.Children)
	}
	// single item
	return json.Unmarshal(data, &u.ItemDto)
}
