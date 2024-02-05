package menu

import (
	"context"
	"strconv"
	"time"

	"fyne.io/systray"
	"github.com/cardinalby/vlc-sync-play/internal/app"
	"github.com/cardinalby/vlc-sync-play/internal/app/static_features"
	"github.com/cardinalby/vlc-sync-play/pkg/tray"
	"github.com/cardinalby/vlc-sync-play/pkg/util/rx"
)

func AddSettingsMenuItems(ctx context.Context, parent *systray.MenuItem, settings *app.Settings) {
	addVlcInstancesMenuItem(ctx, parent, settings.InstancesNumber)
	addNoVideoMenuItem(ctx, parent, settings.NoVideo)
	addReSeekSrcMenuItem(ctx, parent, settings.ReSeekSrc)
	addPollingIntervalMenuItem(ctx, parent, settings.PollingInterval)
	if static_features.ClickPause {
		addClickPauseMenuItem(ctx, parent, settings.ClickPause)
	}
}

func addVlcInstancesMenuItem(ctx context.Context, parent *systray.MenuItem, setting rx.Value[int]) {
	item := tray.GetAddMenuItemFn(parent)(
		"VLC instances",
		"Number of VLC instances to start once file is opened",
	)
	tray.AddOptionsSubMenu(ctx, item, setting, []int{2, 3, 4}, strconv.Itoa)
}

func addPollingIntervalMenuItem(ctx context.Context, parent *systray.MenuItem, setting rx.Value[time.Duration]) {
	item := tray.GetAddMenuItemFn(parent)(
		"Polling interval",
		"How often to poll VLC instances for updates",
	)
	tray.AddOptionsSubMenu(ctx, item, setting, []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		500 * time.Millisecond,
		time.Second,
	}, time.Duration.String)
}

func addNoVideoMenuItem(ctx context.Context, parent *systray.MenuItem, setting rx.Value[bool]) {
	tray.AddBoolOptionMenu(
		ctx,
		parent,
		setting,
		"No video",
		"Open additional VLC instances without video",
	)
}

func addReSeekSrcMenuItem(ctx context.Context, parent *systray.MenuItem, setting rx.Value[bool]) {
	tray.AddBoolOptionMenu(
		ctx,
		parent,
		setting,
		"Re-seek source",
		"Re-seek source player for precise sync",
	)
}

func addClickPauseMenuItem(ctx context.Context, parent *systray.MenuItem, setting rx.Value[bool]) {
	tray.AddBoolOptionMenu(
		ctx,
		parent,
		setting,
		"Click to pause/resume",
		"Click to pause/resume playback",
	)
}
