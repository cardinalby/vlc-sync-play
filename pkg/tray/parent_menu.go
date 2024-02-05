package tray

import "fyne.io/systray"

func GetAddMenuItemFn(
	parent *systray.MenuItem,
) func(title string, tooltip string) *systray.MenuItem {
	if parent == nil {
		return systray.AddMenuItem
	}
	return parent.AddSubMenuItem
}

func GetAddMenuItemCheckboxFn(
	parent *systray.MenuItem,
) func(title string, tooltip string, value bool) *systray.MenuItem {
	if parent == nil {
		return systray.AddMenuItemCheckbox
	}
	return parent.AddSubMenuItemCheckbox
}
