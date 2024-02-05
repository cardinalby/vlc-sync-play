package trayagent

import (
	"context"

	"fyne.io/systray"
	"github.com/cardinalby/vlc-sync-play/assets"
	"github.com/cardinalby/vlc-sync-play/internal/app"
	"github.com/cardinalby/vlc-sync-play/internal/trayagent/menu"
	"github.com/cardinalby/vlc-sync-play/pkg/tray"
	"github.com/cardinalby/vlc-sync-play/pkg/util/logging"
)

type App struct {
	app         *app.App
	appSettings *app.Settings
}

func NewApp(logger logging.Logger) *App {
	return &App{
		app: app.NewApp(logger),
	}
}

func (a *App) Init(settingsPatch app.SettingsPatch) (err error) {
	a.appSettings, err = a.app.Init(settingsPatch)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		<-ctx.Done()
		systray.Quit()
	}()

	syncerErr := make(chan error, 1)
	go func() {
		syncerErr <- a.app.Start(ctx)
		cancel()
	}()

	systray.Run(
		func() {
			a.setupTrayMenu(ctx, cancel)
		},
		func() {
			cancel()
		},
	)

	return <-syncerErr
}

func (a *App) setupTrayMenu(ctx context.Context, onQuit func()) {
	a.setIcon()
	systray.SetTooltip("VLC Sync Play")

	menu.AddSettingsMenuItems(ctx, nil, a.appSettings)

	systray.AddSeparator()

	tray.OnClicked(
		ctx,
		systray.AddMenuItem("Quit", "Quit the app and close players"),
		onQuit,
	)
}

func (a *App) setIcon() {
	icon := assets.GetTrayIcon()
	templateIcon := assets.GetTemplateTrayIcon()
	if templateIcon != nil {
		systray.SetTemplateIcon(templateIcon, templateIcon)
	} else {
		systray.SetIcon(icon)
	}
}
