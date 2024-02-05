package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cardinalby/vlc-sync-play/internal/app/static_features"
	"github.com/cardinalby/vlc-sync-play/pkg/util/logging"
	"github.com/cardinalby/vlc-sync-play/pkg/util/rx"
	typeutil "github.com/cardinalby/vlc-sync-play/pkg/util/type"
	"github.com/kirsle/configdir"
)

const settingsFileName = "settings.json"

type jsonSettings struct {
	InstancesNumber   *int   `json:"instances,omitempty"`
	NoVideo           *bool  `json:"no-video,omitempty"`
	PollingIntervalMs *int64 `json:"interval,omitempty"`
	ClickPause        *bool  `json:"click-pause,omitempty"`
	ReSeekSrc         *bool  `json:"re-seek-src,omitempty"`
}

func (s *jsonSettings) applyToAppSettings(settings *Settings) (updated bool) {
	if s.InstancesNumber != nil {
		settings.InstancesNumber.SetValue(*s.InstancesNumber)
		updated = true
	}
	if s.NoVideo != nil {
		settings.NoVideo.SetValue(*s.NoVideo)
		updated = true
	}
	if s.PollingIntervalMs != nil {
		settings.PollingInterval.SetValue(time.Duration(*s.PollingIntervalMs) * time.Millisecond)
		updated = true
	}
	if s.ClickPause != nil && static_features.ClickPause {
		settings.ClickPause.SetValue(*s.ClickPause)
		updated = true
	}
	if s.ReSeekSrc != nil {
		settings.ReSeekSrc.SetValue(*s.ReSeekSrc)
		updated = true
	}
	return updated
}

func (s *jsonSettings) setFromAppSettings(settings *Settings) {
	s.InstancesNumber = typeutil.Ptr(settings.InstancesNumber.GetValue())
	s.NoVideo = typeutil.Ptr(settings.NoVideo.GetValue())
	s.PollingIntervalMs = typeutil.Ptr(settings.PollingInterval.GetValue().Milliseconds())
	if static_features.ClickPause {
		s.ClickPause = typeutil.Ptr(settings.ClickPause.GetValue())
	}
	s.ReSeekSrc = typeutil.Ptr(settings.ReSeekSrc.GetValue())
}

type SettingsStorage struct {
	configFilePath string
	isFileCreated  bool
	settings       *Settings
	jsonSettings   *jsonSettings // can be nil
	logger         logging.Logger
}

func NewSettingsStorage(settings *Settings, logger logging.Logger) *SettingsStorage {
	return &SettingsStorage{
		configFilePath: filepath.Join(configdir.LocalConfig(Name), settingsFileName),
		settings:       settings,
		jsonSettings:   new(jsonSettings),
		logger:         logger,
	}
}

func (s *SettingsStorage) GetSettings() *Settings {
	return s.settings
}

func (s *SettingsStorage) Load() (loaded bool, err error) {
	s.logger.Info("config location: %s", s.configFilePath)
	fh, err := os.Open(s.configFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	defer func() {
		err = errors.Join(err, fh.Close())
	}()

	decoder := json.NewDecoder(fh)
	if err := decoder.Decode(s.jsonSettings); err != nil {
		return false, err
	}
	return s.jsonSettings.applyToAppSettings(s.settings), nil
}

func (s *SettingsStorage) StartSyncing(ctx context.Context) error {
	observers := rx.Observers{}
	defer observers.UnsubscribeAll()
	syncErrCh := make(chan error, 1)

	observers = append(observers, s.settings.InstancesNumber.Subscribe(func(v int) {
		s.jsonSettings.InstancesNumber = &v
		s.saveJsonSettingsWithErrChan(syncErrCh)
	}))
	observers = append(observers, s.settings.NoVideo.Subscribe(func(v bool) {
		s.jsonSettings.NoVideo = &v
		s.saveJsonSettingsWithErrChan(syncErrCh)
	}))
	observers = append(observers, s.settings.PollingInterval.Subscribe(func(v time.Duration) {
		s.jsonSettings.PollingIntervalMs = typeutil.Ptr(v.Milliseconds())
		s.saveJsonSettingsWithErrChan(syncErrCh)
	}))
	if static_features.ClickPause {
		observers = append(observers, s.settings.ClickPause.Subscribe(func(v bool) {
			s.jsonSettings.ClickPause = &v
			s.saveJsonSettingsWithErrChan(syncErrCh)
		}))
	}
	observers = append(observers, s.settings.ReSeekSrc.Subscribe(func(v bool) {
		s.jsonSettings.ReSeekSrc = &v
		s.saveJsonSettingsWithErrChan(syncErrCh)
	}))

	select {
	case <-ctx.Done():
		return nil
	case err := <-syncErrCh:
		return err
	}
}

func (s *SettingsStorage) Save() error {
	if s.jsonSettings == nil {
		s.jsonSettings.setFromAppSettings(s.settings)
	}
	return s.saveJsonSettings()
}

func (s *SettingsStorage) saveJsonSettings() (err error) {
	var f *os.File
	defer func() {
		if f != nil {
			err = errors.Join(err, f.Close())
		}
	}()

	if f, err = s.openFileForWrite(); err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	encoder := json.NewEncoder(f)
	return encoder.Encode(&s.jsonSettings)
}

func (s *SettingsStorage) saveJsonSettingsWithErrChan(errChan chan<- error) {
	if err := s.saveJsonSettings(); err != nil {
		select {
		case errChan <- err:
		default:
		}
	}
}

func (s *SettingsStorage) openFileForWrite() (*os.File, error) {
	openFile := func() (*os.File, error) {
		return os.OpenFile(s.configFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	}

	f, err := openFile()
	if err == nil {
		return f, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}

	if err := configdir.MakePath(filepath.Dir(s.configFilePath)); err != nil {
		return nil, fmt.Errorf("error creating settings dir: %w", err)
	}

	return openFile()
}
