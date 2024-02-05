package tray

import (
	"context"

	"fyne.io/systray"
	"github.com/cardinalby/vlc-sync-play/pkg/util/rx"
)

func AddBoolOptionMenu(
	ctx context.Context,
	parent *systray.MenuItem,
	setting rx.Value[bool],
	title string,
	tooltip string,
) {
	menuItem := GetAddMenuItemCheckboxFn(parent)(title, tooltip, setting.GetValue())
	OnClicked(
		ctx,
		menuItem,
		func() {
			val := !setting.GetValue()
			setting.SetValue(val)
			if val {
				menuItem.Check()
			} else {
				menuItem.Uncheck()
			}
		},
	)
}
