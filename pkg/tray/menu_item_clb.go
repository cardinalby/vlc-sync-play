package tray

import (
	"context"

	"fyne.io/systray"
)

func OnClicked(
	ctx context.Context,
	menuItem *systray.MenuItem,
	callback func(),
) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-menuItem.ClickedCh:
				callback()
			}
		}
	}()
}
