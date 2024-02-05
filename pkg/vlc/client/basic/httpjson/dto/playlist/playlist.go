package playlist_dto

type ItemType string

const ItemTypeNode ItemType = "node"
const ItemTypeLeaf ItemType = "leaf"

type ItemDto struct {
	Type     ItemType  `json:"type"`
	Name     string    `json:"name"`
	Uri      string    `json:"uri"`
	Current  string    `json:"current"`
	Children []ItemDto `json:"children"`
}

func (pl ItemDto) GetCurrent() (ItemDto, bool) {
	if pl.Type == ItemTypeLeaf {
		return pl, pl.Current != ""
	}
	for _, child := range pl.Children {
		if curr, ok := child.GetCurrent(); ok {
			return curr, true
		}
	}

	return pl, false
}
