package syncer

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cardinalby/vlc-sync-play/pkg/util/logging"
	timeutil "github.com/cardinalby/vlc-sync-play/pkg/util/time"
	typeutil "github.com/cardinalby/vlc-sync-play/pkg/util/type"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/extended"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/timings"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/instance"
	"golang.org/x/sync/errgroup"
)

type Syncer struct {
	syncingMu              sync.Mutex
	lastPlayerId           uint
	players                *players
	settings               Settings
	lastSyncedAt           time.Time
	isWaitingForUpdateFrom *player // pause sync until got update from this player
	isStarted              atomic.Bool
	instanceLauncher       instance.Launcher
	logger                 logging.Logger
}

func NewSyncer(
	masterInstance instance.Instance,
	settings Settings,
	instanceLauncher instance.Launcher,
	logger logging.Logger,
) *Syncer {
	return &Syncer{
		players: newPlayers(newPlayer(
			masterInstance,
			getPlayerSettings(settings),
			logger,
			0,
		)),
		settings:         settings,
		instanceLauncher: instanceLauncher,
		logger:           logger,
		lastPlayerId:     0,
	}
}

func (s *Syncer) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)

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
			for {
				if _, shouldRepeat := pl.SendCmdGroup(ctx, commands); !shouldRepeat || ctx.Err() != nil {
					return
				}
				if err := timeutil.SleepCtx(ctx, timings.CommandsRepeatInterval); err != nil {
					return
				}
			}
		}()
		return true
	})
	waitGr.Wait()

	s.syncPlayersPosition(ctx, commands.Seek.Value, nil)
}

func (s *Syncer) onUpdate(ctx context.Context, plUpdate *playerUpdate) error {
	s.syncingMu.Lock()
	defer s.syncingMu.Unlock()

	if s.isWaitingForUpdateFrom != nil && s.isWaitingForUpdateFrom != plUpdate.player {
		s.logger.Info("Skipping [%d] update because is waiting only from %d",
			plUpdate.player.id, s.isWaitingForUpdateFrom.id)
		return nil
	}
	if plUpdate.update.Status.Moment.Center().Before(s.lastSyncedAt) {
		s.logger.Info("Skipping [%d] update from %v old sync iteration: pos %v",
			plUpdate.player.id, s.lastSyncedAt.Sub(plUpdate.update.Status.Moment.Max), plUpdate.update.Status.Position)
		return nil
	}

	if plUpdate.update.ChangedProps.HasFileURI() &&
		plUpdate.update.Status.State != basic.PlaybackStateStopped {
		// File opened
		s.isWaitingForUpdateFrom = plUpdate.player
		missingInstancesNumber := s.settings.GetInstancesNumber().GetValue() - s.players.Len()
		if missingInstancesNumber > 0 {
			return s.launchAdditionalInstances(
				ctx,
				plUpdate.update.Status.FileURI,
				missingInstancesNumber,
			)
		}
	} else {
		s.isWaitingForUpdateFrom = nil
	}

	if !plUpdate.update.IsNatural {
		s.syncPlayers(ctx, plUpdate)
	}
	return nil
}

func (s *Syncer) launchAdditionalInstances(
	ctx context.Context,
	openFile string,
	missingInstancesNumber int,
) error {
	options := instance.LaunchOptions{
		NoVideo: s.settings.GetNoVideo().GetValue(),
	}
	if openFile != "" {
		options.FilePaths = []string{openFile}
	}
	errGr := errgroup.Group{}

	for i := 0; i < missingInstancesNumber; i++ {
		errGr.Go(func() error {
			newInstance, err := s.instanceLauncher(ctx, options)
			if err != nil {
				return fmt.Errorf("failed to create new instance: %w", err)
			}
			s.lastPlayerId++
			newPlayer := newPlayer(
				newInstance,
				getPlayerSettings(s.settings),
				s.logger,
				s.lastPlayerId,
			)
			s.players.Add(newPlayer)
			return nil
		})
	}
	return errGr.Wait()
}

func (s *Syncer) syncPlayers(
	ctx context.Context,
	srcUpdate *playerUpdate,
) {
	s.logger.Info("-- Syncing caused by %d update: %s", srcUpdate.player.id, srcUpdate.update)
	commands := srcUpdate.GetSyncCommands()
	s.syncOtherPlayersNoSeek(ctx, srcUpdate, commands)
	if commands.Seek.HasValue {
		var skipPlayer *player
		if !s.settings.GetReSeekSrc().GetValue() {
			skipPlayer = srcUpdate.player
		}
		s.syncPlayersPosition(ctx, commands.Seek.Value, skipPlayer)
	}
	s.lastSyncedAt = time.Now()
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
			s.logger.Info("P[%d]: additional sync [%s] -> [%s]", pl.id, srcUpdate.update, dstUpdate)
			dstCommands = srcUpdate.player.client.state.GetSyncCommands(
				srcUpdate.update.ChangedProps.Union(dstUpdate.ChangedProps),
			)
		}

		waitGr.Add(1)
		go func() {
			defer waitGr.Done()
			for {
				if _, shouldRepeat := pl.SendCmdGroup(ctx, dstCommands); !shouldRepeat || ctx.Err() != nil {
					return
				}
				if err := timeutil.SleepCtx(ctx, timings.CommandsRepeatInterval); err != nil {
					return
				}
			}
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

	allSeeked := atomic.Bool{}
	for !allSeeked.Load() {
		waitGr := sync.WaitGroup{}
		allSeeked.Store(true)

		s.players.Iterate(func(pl *player) bool {
			if pl == skipPlayer {
				return true
			}
			waitGr.Add(1)
			go func() {
				defer waitGr.Done()
				if _, shouldRepeat := pl.SendCmdGroup(ctx, commands); shouldRepeat {
					allSeeked.Store(false)
				}
			}()
			return true
		})
		waitGr.Wait()
	}
}
