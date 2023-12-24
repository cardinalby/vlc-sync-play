package interactive

import (
	"context"

	"github.com/cardinalby/vlc-sync-play/internal/app"
	"github.com/cardinalby/vlc-sync-play/internal/cli/interactive/settings"
	"github.com/cardinalby/vlc-sync-play/pkg/util/logging"
	"github.com/rivo/tview"
	"golang.org/x/sync/errgroup"
)

type App struct {
	app      *app.App
	tviewApp *tview.Application
}

func NewApp(logger logging.Logger) *App {
	return &App{
		app:      app.NewApp(logger),
		tviewApp: tview.NewApplication().EnableMouse(true),
	}
}

func (a *App) Init(settingsPatch app.SettingsPatch) error {
	appSettings, err := a.app.Init(settingsPatch)
	if err != nil {
		return err
	}
	a.tviewApp.SetRoot(settings.BuildRoot(appSettings), false)
	return nil
}

func (a *App) Start(ctx context.Context) error {
	parentCtx := ctx
	ctx, cancel := context.WithCancel(parentCtx)
	errGroup, ctx := errgroup.WithContext(ctx)

	errGroup.Go(func() error {
		defer cancel()
		return a.app.Start(ctx)
	})
	errGroup.Go(func() error {
		defer cancel()
		return a.tviewApp.Run()
	})
	errGroup.Go(func() error {
		<-ctx.Done()
		a.tviewApp.Stop()
		return nil
	})

	return errGroup.Wait()
}
