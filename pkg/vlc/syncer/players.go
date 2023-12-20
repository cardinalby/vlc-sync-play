package syncer

import (
	"context"
	"errors"
	"sync"

	"golang.org/x/sync/errgroup"
)

type players struct {
	items        []*player
	mu           sync.RWMutex
	waitCtx      context.Context
	waitErrGroup *errgroup.Group
	updatesChan  chan<- playerUpdate
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
		item.StartWaiting(s.waitCtx, s.waitErrGroup.Go, s.updatesChan)
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

func (s *players) WaitAndPoll(ctx context.Context, updatesReceiver chan<- playerUpdate) error {
	s.mu.Lock()
	if s.waitErrGroup != nil {
		return errors.New("already waiting")
	}

	s.waitErrGroup, s.waitCtx = errgroup.WithContext(ctx)
	s.updatesChan = updatesReceiver

	for _, pl := range s.items {
		pl.StartWaiting(s.waitCtx, s.waitErrGroup.Go, s.updatesChan)
	}
	s.mu.Unlock()

	err := s.waitErrGroup.Wait()

	s.mu.Lock()
	s.waitCtx = nil
	s.waitErrGroup = nil
	close(s.updatesChan)
	s.updatesChan = nil
	s.mu.Unlock()

	return err
}
