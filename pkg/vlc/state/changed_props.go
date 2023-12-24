package state

import "strings"

type ChangedProps [4]bool

func (cp *ChangedProps) HasFileURI() bool {
	return cp[0]
}

func (cp *ChangedProps) HasPosition() bool {
	return cp[1]
}

func (cp *ChangedProps) HasState() bool {
	return cp[2]
}

func (cp *ChangedProps) HasRate() bool {
	return cp[3]
}

func (cp *ChangedProps) HasAny() bool {
	for _, v := range cp {
		if v {
			return true
		}
	}
	return false
}

func (cp *ChangedProps) HasAll() bool {
	for _, v := range cp {
		if !v {
			return false
		}
	}
	return true
}

func (cp *ChangedProps) SetFileURI(value bool) {
	cp[0] = value
}

func (cp *ChangedProps) SetPosition(value bool) {
	cp[1] = value
}

func (cp *ChangedProps) SetState(value bool) {
	cp[2] = value
}

func (cp *ChangedProps) SetRate(value bool) {
	cp[3] = value
}

func (cp *ChangedProps) SetAll(value bool) {
	for i := range cp {
		cp[i] = value
	}
}

func (cp *ChangedProps) String() string {
	names := make([]string, 0, 4)
	if cp.HasFileURI() {
		names = append(names, "FileURI")
	}
	if cp.HasPosition() {
		names = append(names, "Position")
	}
	if cp.HasState() {
		names = append(names, "State")
	}
	if cp.HasRate() {
		names = append(names, "Rate")
	}
	return strings.Join(names, ", ")
}

func (cp *ChangedProps) Union(another ChangedProps) ChangedProps {
	for i := 0; i < len(another); i++ {
		another[i] = cp[i] || another[i]
	}
	return another
}

func (cp *ChangedProps) Includes(another ChangedProps) bool {
	for i := 0; i < len(another); i++ {
		if another[i] && !cp[i] {
			return false
		}
	}
	return true
}
