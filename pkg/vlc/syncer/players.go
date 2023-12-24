package syncer

import (
	"context"
	"errors"
	"sync"

	"github.com/cardinalby/vlc-sync-play/pkg/vlc/instance"
	"golang.org/x/sync/errgroup"
)

type players struct {
	items         []*player
	mu            sync.RWMutex
	waitCtx       context.Context
	waitErrGroup  *errgroup.Group
	onUpdate      func(playerUpdate)
	onPlayerEvent func(playerEvent)
}

func newPlayers(items ...*player) *players {
	return &players{
		items: items,
		mu:    sync.RWMutex{},
	}
}

func (s *players) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

func (s *players) Add(item *player) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, item)
	if s.waitErrGroup != nil {
		onEvent := func(event instance.StdErrEvent) {
			s.onPlayerEvent(playerEvent{event: event, player: item})
		}
		item.StartWaiting(s.waitCtx, s.waitErrGroup.Go, s.onUpdate, onEvent)
	}
}

func (s *players) Iterate(yield func(*player) (next bool)) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := 0; i < len(s.items); i++ {
		if !yield(s.items[i]) {
			break
		}
	}
}

func (s *players) WaitAndPoll(
	ctx context.Context,
	onUpdate func(update playerUpdate),
	onPlayerEvent func(event playerEvent),
) error {
	s.mu.Lock()
	if s.waitErrGroup != nil {
		return errors.New("already waiting")
	}

	s.onUpdate = onUpdate
	s.onPlayerEvent = onPlayerEvent
	s.waitErrGroup, s.waitCtx = errgroup.WithContext(ctx)

	for _, pl := range s.items {
		onEvent := func(event instance.StdErrEvent) {
			onPlayerEvent(playerEvent{event: event, player: pl})
		}
		pl.StartWaiting(s.waitCtx, s.waitErrGroup.Go, onUpdate, onEvent)
	}
	s.mu.Unlock()

	err := s.waitErrGroup.Wait()

	s.mu.Lock()
	s.waitCtx = nil
	s.waitErrGroup = nil
	s.mu.Unlock()

	return err
}
