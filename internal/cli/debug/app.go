package debug

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/cardinalby/vlc-sync-play/internal/app"
	"github.com/cardinalby/vlc-sync-play/pkg/util/logging"
)

type App struct {
	app *app.App
}

func NewApp(logger logging.Logger) *App {
	return &App{
		app: app.NewApp(logger),
	}
}

func (a *App) Init(settingsPatch app.SettingsPatch) error {
	_, err := a.app.Init(settingsPatch)
	return err
}

func (a *App) Start(ctx context.Context) error {
	return a.app.Start(a.getAppContext(ctx))
}

func (a *App) getAppContext(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	call := make(chan os.Signal, 1)
	signal.Notify(call, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-call
		cancel()
	}()
	return ctx
}
