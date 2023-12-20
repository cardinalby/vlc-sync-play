package syncer

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	timeutil "github.com/cardinalby/vlc-sync-play/pkg/util/time"
	typeutil "github.com/cardinalby/vlc-sync-play/pkg/util/type"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/extended"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/instance"
	"golang.org/x/sync/errgroup"
)

const commandsRepeatInterval = 50 * time.Millisecond

type Syncer struct {
	players               *players
	lastSyncedAt          time.Time
	targetInstancesNumber int
	pollingInterval       time.Duration
	isStarted             atomic.Bool
	instanceLauncher      instance.Launcher
}

func NewSyncer(
	masterInstance instance.Instance,
	pollingInterval time.Duration,
	targetInstancesNumber int,
	instanceLauncher instance.Launcher,
) *Syncer {
	return &Syncer{
		players:               newPlayers(newPlayer(masterInstance, pollingInterval)),
		pollingInterval:       pollingInterval,
		targetInstancesNumber: targetInstancesNumber,
		instanceLauncher:      instanceLauncher,
	}
}

func (s *Syncer) Start(ctx context.Context) error {
	errGr, ctx := errgroup.WithContext(ctx)
	updatesChan := make(chan playerUpdate)
	errGr.Go(func() error {
		return s.players.WaitAndPoll(ctx, updatesChan)
	})
	errGr.Go(func() error {
		return s.StartPollingLoop(ctx, updatesChan)
	})
	return errGr.Wait()
}

func (s *Syncer) StartPollingLoop(ctx context.Context, updatesChan <-chan playerUpdate) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case plUpdate, ok := <-updatesChan:
			if !ok {
				return ctx.Err()
			}
			if s.players.Len() == 1 {
				if plUpdate.update.ChangedProps.HasFileURI() {
					if err := s.startAdditionalInstances(); err != nil {
						return fmt.Errorf("failed to init new instances: %w", err)
					}
					plUpdate.update.ChangedProps.SetAll(true)
					s.syncPlayers(ctx, &plUpdate)
				}
			} else {
				s.onUpdate(ctx, &plUpdate)
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
		}
	}
}

func (s *Syncer) startAdditionalInstances() error {
	errGr := errgroup.Group{}
	for i := 0; i < s.targetInstancesNumber-1; i++ {
		errGr.Go(func() error {
			newInstance, err := s.instanceLauncher(nil, true)
			if err != nil {
				return fmt.Errorf("failed to create new instance: %w", err)
			}
			newPlayer := newPlayer(newInstance, s.pollingInterval)
			s.players.Add(newPlayer)
			return nil
		})
	}
	return errGr.Wait()
}

func (s *Syncer) onUpdate(ctx context.Context, update *playerUpdate) {
	if update.update.Status.Moment.Max.Before(s.lastSyncedAt) {
		fmt.Printf("Skipping [%d] update from old sync iteration: %s (actual: %s)\n",
			update.player.client.index, update.update.Status.Moment.Max, s.lastSyncedAt)
		return
	}
	if !update.update.IsNatural {
		s.syncPlayers(ctx, update)
	}
}

func (s *Syncer) syncPlayers(ctx context.Context, srcUpdate *playerUpdate) {
	fmt.Printf("-- Syncing caused by %d update: %s\n", srcUpdate.player.client.index, srcUpdate.update.String())
	commands := srcUpdate.GetSyncCommands()
	s.syncOtherPlayersNoSeek(ctx, srcUpdate, commands)
	if commands.Seek.HasValue {
		s.syncAllPlayersPosition(ctx, commands.Seek.Value)
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

		if err == nil && dstUpdate.ChangedProps != srcUpdate.update.ChangedProps {
			fmt.Printf("Player %d: additional sync [%s] -> [%s]\n", pl.client.index, srcUpdate.update.String(), dstUpdate.String())
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
				err := timeutil.SleepCtx(ctx, commandsRepeatInterval)
				if err != nil {
					return
				}
			}
		}()
		return true
	})
	waitGr.Wait()
}

func (s *Syncer) syncAllPlayersPosition(
	ctx context.Context,
	positionGetter extended.ExpectedPositionGetter,
) {
	commands := extended.CmdGroup{
		Seek: typeutil.NewOptional(positionGetter),
	}

	allSeeked := atomic.Bool{}
	for !allSeeked.Load() {
		waitGr := sync.WaitGroup{}
		allSeeked.Store(true)

		s.players.Iterate(func(pl *player) bool {
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
