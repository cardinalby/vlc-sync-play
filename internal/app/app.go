package app

import (
	"context"
	"errors"
	"fmt"

	errutil "github.com/cardinalby/vlc-sync-play/pkg/util/err"
	"github.com/cardinalby/vlc-sync-play/pkg/util/logging"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/instance"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/syncer"
	"golang.org/x/sync/errgroup"
)

const Name = "vlc-sync-play"

var ErrSettingsPatchIsInvalid = errors.New("settings patch is invalid")

type App struct {
	logger          logging.Logger
	settingsStorage *SettingsStorage
}

func NewApp(
	logger logging.Logger,
) *App {
	return &App{
		logger: logger,
	}
}

func (a *App) Init(settingsPatch SettingsPatch) (settings *Settings, err error) {
	if a.settingsStorage, err = a.createSettingsStorage(settingsPatch); err != nil {
		return nil, err
	}
	return a.settingsStorage.GetSettings(), nil
}

func (a *App) Start(ctx context.Context) (err error) {
	if a.settingsStorage == nil {
		return fmt.Errorf("app is not initialized")
	}
	settings := a.settingsStorage.GetSettings()

	instanceLauncher := instance.GetLauncher(settings.VlcPath, settings.ApiProtocol, a.logger)
	masterInstance, err := instanceLauncher(ctx, instance.LaunchOptions{
		FilePaths: settings.FilePaths,
	})
	if err != nil {
		return err
	}
	playersSyncer := syncer.NewSyncer(
		masterInstance,
		settings,
		instanceLauncher,
		a.logger,
	)

	settingsSyncCtx, settingsSyncCtxCancel := context.WithCancel(ctx)
	errGroup, ctx := errgroup.WithContext(ctx)

	errGroup.Go(func() error {
		defer settingsSyncCtxCancel()
		err := playersSyncer.Start(ctx)
		if errors.Is(err, ctx.Err()) || errors.Is(err, syncer.ErrInstanceFinished) {
			return nil
		}
		return err
	})
	errGroup.Go(func() error {
		return a.settingsStorage.StartSyncing(settingsSyncCtx)
	})
	return errGroup.Wait()
}

func (a *App) createSettingsStorage(settingsPatch SettingsPatch) (*SettingsStorage, error) {
	settings := NewSettings()
	settings.SetDefaults()

	settingsStorage := NewSettingsStorage(settings)
	if _, loadErr := settingsStorage.Load(); loadErr != nil {
		a.logger.Err("error loading settings: %s", loadErr.Error())
		if saveErr := settingsStorage.Save(); saveErr != nil {
			return nil, errutil.Join(loadErr, saveErr)
		}
	} else {
		if validateErr := settings.Validate(); validateErr != nil {
			a.logger.Err("error validating loaded settings: %s", validateErr.Error())
			settings.SetDefaults()
			if saveErr := settingsStorage.Save(); saveErr != nil {
				return nil, errutil.Join(validateErr, saveErr)
			}
		}
	}

	if settingsPatch.ApplyToSettings(settings) {
		if err := settings.Validate(); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrSettingsPatchIsInvalid, err.Error())
		}
	}
	if settings.VlcPath == "" {
		if defaultPath, err := instance.GetDefaultVlcBinPath(); err == nil {
			settings.VlcPath = defaultPath
		} else {
			return nil, fmt.Errorf("error getting default VLC path: %w", err)
		}
	}

	return settingsStorage, nil
}
