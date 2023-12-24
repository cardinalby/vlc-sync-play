package app

import (
	"errors"
	"time"

	"github.com/cardinalby/vlc-sync-play/pkg/util/rx"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic/protocols"
)

type SettingsPatch interface {
	ApplyToSettings(s *Settings) (updated bool)
}

type Settings struct {
	ApiProtocol     protocols.ApiProtocol
	VlcPath         string
	FilePaths       []string
	InstancesNumber rx.Value[int]
	NoVideo         rx.Value[bool]
	PollingInterval rx.Value[time.Duration]
	ClickPause      rx.Value[bool]
	ReSeekSrc       rx.Value[bool]
}

func NewSettings() *Settings {
	return &Settings{
		ApiProtocol:     protocols.ApiProtocolHttpJson,
		InstancesNumber: rx.NewValue[int](0),
		NoVideo:         rx.NewValue[bool](false),
		PollingInterval: rx.NewValue[time.Duration](0),
		ClickPause:      rx.NewValue[bool](false),
		ReSeekSrc:       rx.NewValue[bool](false),
	}
}

func (s *Settings) SetDefaults() {
	s.ApiProtocol = protocols.ApiProtocolHttpJson
	s.InstancesNumber.SetValue(2)
	s.NoVideo.SetValue(false)
	s.PollingInterval.SetValue(100 * time.Millisecond)
	s.ClickPause.SetValue(true)
	s.ReSeekSrc.SetValue(true)
}

func (s *Settings) GetPollingInterval() rx.Observable[time.Duration] {
	return s.PollingInterval
}

func (s *Settings) GetInstancesNumber() rx.Observable[int] {
	return s.InstancesNumber
}

func (s *Settings) GetNoVideo() rx.Observable[bool] {
	return s.NoVideo
}

func (s *Settings) GetClickPause() rx.Observable[bool] {
	return s.ClickPause
}

func (s *Settings) GetReSeekSrc() rx.Observable[bool] {
	return s.ReSeekSrc
}

func (s *Settings) Validate() error {
	if err := s.ApiProtocol.Validate(); err != nil {
		return err
	}
	if s.InstancesNumber.GetValue() < 2 {
		return errors.New("instances number should be at least 2")
	}
	if s.PollingInterval.GetValue() < 0 {
		return errors.New("polling interval should be positive")
	}
	return nil
}
