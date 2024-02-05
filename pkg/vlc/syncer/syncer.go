package syncer

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cardinalby/vlc-sync-play/pkg/util/logging"
	typeutil "github.com/cardinalby/vlc-sync-play/pkg/util/type"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/extended"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/extended/repetition"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/timings"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/instance"
	"golang.org/x/sync/errgroup"
)

type Syncer struct {
	syncingMu                    sync.Mutex
	players                      *players
	settings                     Settings
	followersSkipUpdatesDuration time.Duration
	state                        State
	isStarted                    atomic.Bool
	instanceLauncher             instance.Launcher
	logger                       logging.Logger
}

func NewSyncer(
	settings Settings,
	instanceLauncher instance.Launcher,
	logger logging.Logger,
) *Syncer {
	return &Syncer{
		players:  newPlayers(),
		settings: settings,
		followersSkipUpdatesDuration: timings.GetFollowerUpdatesIgnoreDuration(
			settings.GetPollingInterval().GetValue(),
		),
		state:            NewState(),
		instanceLauncher: instanceLauncher,
		logger:           logger,
	}
}

func (s *Syncer) Start(ctx context.Context, initFileURI string) error {
	if err := s.launchInstances(ctx, initFileURI, 1); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)

	defer s.settings.GetInstancesNumber().Subscribe(func(value int) {
		s.launchMissingInstances(ctx, value)
	}).Unsubscribe()

	defer s.settings.GetPollingInterval().Subscribe(func(value time.Duration) {
		s.syncingMu.Lock()
		defer s.syncingMu.Unlock()
		s.followersSkipUpdatesDuration = timings.GetFollowerUpdatesIgnoreDuration(value)
	})

	return s.players.WaitAndPoll(
		ctx,
		func(update playerUpdate) {
			if err := s.onUpdate(ctx, &update); err != nil {
				cancel()
			}
		},
		func(event playerEvent) {
			if err := s.onEvent(ctx, event); err != nil {
				cancel()
			}
		},
		s.onFinished,
	)
}

func (s *Syncer) onEvent(ctx context.Context, event playerEvent) error {
	switch event.event {
	case instance.StderrEventMouse1Click:
		commands := event.player.client.state.GetPauseOrResumeCommand()
		s.sendAllPlayersCommands(ctx, commands)
	}
	return nil
}

func (s *Syncer) onFinished(pl *player) {
	s.logger.Info("P[%d]: finished", pl.GetID())
	s.syncingMu.Lock()
	defer s.syncingMu.Unlock()
	if s.state.lastSyncedFromID == pl.GetID() {
		s.state.lastSyncedFromID = instance.IDNone
	}
}

func (s *Syncer) sendAllPlayersCommands(
	ctx context.Context,
	commands extended.CmdGroup,
) {
	noSeekCommands := commands
	noSeekCommands.Seek.Reset()

	waitGr := sync.WaitGroup{}
	s.players.Iterate(func(pl *player) bool {
		waitGr.Add(1)
		go func() {
			defer waitGr.Done()
			_, _ = pl.SendCmdGroup(ctx, commands, repetition.WithInterval(timings.CommandsRepeatInterval))
		}()
		return true
	})
	waitGr.Wait()

	s.syncPlayersPosition(ctx, commands.Seek.Value, nil)
}

func (s *Syncer) onUpdate(ctx context.Context, plUpdate *playerUpdate) error {
	s.syncingMu.Lock()
	defer s.syncingMu.Unlock()

	if canAccept, getReason := s.checkCanAcceptUpdate(plUpdate); !canAccept {
		s.logger.Info(getReason())
		return nil
	}

	s.state.fileURI.SetValue(plUpdate.update.Status.FileURI)
	s.state.lastSyncedFromID = plUpdate.player.GetID()

	if plUpdate.update.ChangedProps.HasFileURI() &&
		plUpdate.update.Status.State != basic.PlaybackStateStopped {
		s.onFileOpened(ctx, plUpdate.player.GetID())
		return nil
	}

	if !plUpdate.update.IsNatural {
		s.syncPlayers(ctx, plUpdate)
	}
	return nil
}

func (s *Syncer) checkCanAcceptUpdate(plUpdate *playerUpdate) (canAccept bool, getReason func() string) {
	plID := plUpdate.player.GetID()
	if plID == s.state.lastSyncedFromID || s.state.lastSyncedFromID == instance.IDNone {
		return true, nil
	}

	if plUpdate.update.Status.Moment.Center().Before(s.state.acceptFollowerUpdatesAfter) {
		return false, func() string {
			return fmt.Sprintf("Skipping [%d] update from %v old sync iteration: pos %v",
				plUpdate.player.GetID(),
				s.state.acceptFollowerUpdatesAfter.Sub(plUpdate.update.Status.Moment.Max),
				plUpdate.update.Status.Position,
			)
		}
	}
	return true, nil
}

func (s *Syncer) launchMissingInstances(ctx context.Context, targetInstancesNumber int) {
	missing := targetInstancesNumber - s.players.Len()
	if missing <= 0 {
		return
	}

	if fileURI := s.state.fileURI.GetValue(); fileURI != "" {
		_ = s.launchInstances(ctx, fileURI, missing)
	} else {
		defer s.state.fileURI.Subscribe(func(fileURI string) {
			_ = s.launchInstances(ctx, fileURI, missing)
		}).Unsubscribe()
	}
	return
}

func (s *Syncer) launchInstances(
	ctx context.Context,
	fileURI string,
	missingInstancesNumber int,
) error {
	options := instance.LaunchOptions{
		// First instance will be launched with video
		NoVideo: s.players.Len() > 0 && s.settings.GetNoVideo().GetValue(),
		FileURI: typeutil.Optional[string]{
			HasValue: fileURI != "",
			Value:    fileURI,
		},
	}
	errGr := errgroup.Group{}

	for i := 0; i < missingInstancesNumber; i++ {
		errGr.Go(func() error {
			s.logger.Info("Launching new instance")

			newInstance, err := s.instanceLauncher.Launch(ctx, options)
			if err != nil {
				return fmt.Errorf("failed to create new instance: %w", err)
			}
			s.players.Add(newPlayer(
				newInstance,
				getPlayerSettings(s.settings),
				s.logger,
			))
			return nil
		})
	}
	return errGr.Wait()
}

func (s *Syncer) onFileOpened(ctx context.Context, srcPlayerID uint) {
	// The source player may auto-seek and will send the next update with other properties
	s.state.acceptFollowerUpdatesAfter = time.Now().Add(max(
		s.followersSkipUpdatesDuration,
		timings.WaitForAutoSeekAfterFileOpenedDuration,
	))
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.launchMissingInstances(
			ctx,
			s.settings.GetInstancesNumber().GetValue(),
		)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.players.Iterate(func(pl *player) bool {
			if pl.GetID() != srcPlayerID {
				_, _ = pl.SendCmdGroup(
					ctx,
					extended.CmdGroup{
						OpenFile: typeutil.NewOptional(s.state.fileURI.GetValue()),
					},
					repetition.WithInterval(timings.CommandsRepeatInterval),
				)
			}
			return true
		})
	}()

	wg.Wait()
}

func (s *Syncer) syncPlayers(
	ctx context.Context,
	srcUpdate *playerUpdate,
) {
	s.logger.Info("-- Syncing caused by %d update: %s", srcUpdate.player.GetID(), srcUpdate.update.String())
	commands := srcUpdate.GetSyncCommands()
	s.syncOtherPlayersNoSeek(ctx, srcUpdate, commands)
	if commands.Seek.HasValue && s.players.Len() > 1 {
		var skipPlayer *player
		if !s.settings.GetReSeekSrc().GetValue() {
			skipPlayer = srcUpdate.player
		}
		s.syncPlayersPosition(ctx, commands.Seek.Value, skipPlayer)
	}
	s.state.acceptFollowerUpdatesAfter = time.Now().Add(s.followersSkipUpdatesDuration)
}

func (s *Syncer) syncOtherPlayersNoSeek(
	ctx context.Context,
	srcUpdate *playerUpdate,
	commands extended.CmdGroup,
) {
	commands.Seek.Reset()
	srcUpdate.update.ChangedProps.SetPosition(false)
	if !commands.HasAny() {
		return
	}

	waitGr := sync.WaitGroup{}

	s.players.Iterate(func(pl *player) bool {
		dstCommands := commands
		if srcUpdate.player == pl {
			return true
		}

		// Check if additional props sync required
		dstUpdate, err := pl.client.state.GetUpdate(&srcUpdate.update.Status)
		dstUpdate.ChangedProps.SetPosition(false)

		if err == nil && !srcUpdate.update.ChangedProps.Includes(dstUpdate.ChangedProps) {
			s.logger.Info("P[%d]: additional sync [%s] -> [%s]", pl.GetID(), srcUpdate.update, dstUpdate)
			dstCommands = srcUpdate.player.client.state.GetSyncCommands(
				srcUpdate.update.ChangedProps.Union(dstUpdate.ChangedProps),
			)
		}

		waitGr.Add(1)
		go func() {
			defer waitGr.Done()
			_, _ = pl.SendCmdGroup(ctx, dstCommands, repetition.WithInterval(timings.CommandsRepeatInterval))
		}()
		return true
	})
	waitGr.Wait()
}

func (s *Syncer) syncPlayersPosition(
	ctx context.Context,
	positionGetter extended.ExpectedPositionGetter,
	skipPlayer *player,
) {
	commands := extended.CmdGroup{
		Seek: typeutil.NewOptional(positionGetter),
	}

	for {
		wg := sync.WaitGroup{}
		var hasNotRecoverableErr, hasRecoverableErr atomic.Bool
		hasNotRecoverableErr.Store(false)
		hasRecoverableErr.Store(false)

		s.players.Iterate(func(pl *player) bool {
			if pl == skipPlayer {
				return true
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				if _, err := pl.SendCmdGroup(ctx, commands, repetition.Single()); err != nil {
					s.logger.Err("Failed to sync position: %s", err.Error())
					if pl.IsRecoverableErr(err) {
						hasRecoverableErr.Store(true)
					} else {
						hasNotRecoverableErr.Store(true)
					}
				}
			}()
			return true
		})
		wg.Wait()
		if ctx.Err() != nil || !hasRecoverableErr.Load() || hasNotRecoverableErr.Load() {
			return
		}
	}
}
