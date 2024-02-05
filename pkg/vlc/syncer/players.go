package syncer

import (
	"context"
	"errors"
	"sync"

	"github.com/cardinalby/vlc-sync-play/pkg/vlc/instance"
	"golang.org/x/sync/errgroup"
)

type players struct {
	items        map[*player]struct{}
	mu           sync.RWMutex
	waitCtx      context.Context
	waitErrGroup *errgroup.Group
	onUpdate     func(playerUpdate)
	onEvent      func(playerEvent)
	onFinish     func(*player)
}

func newPlayers() *players {
	return &players{
		items: make(map[*player]struct{}),
		mu:    sync.RWMutex{},
	}
}

func (pls *players) Len() int {
	pls.mu.RLock()
	defer pls.mu.RUnlock()
	return len(pls.items)
}

func (pls *players) Add(item *player) {
	pls.mu.Lock()
	defer pls.mu.Unlock()
	pls.items[item] = struct{}{}
	if pls.waitErrGroup != nil {
		pls.startWaitingForPlayer(item)
	}
}

func (pls *players) Iterate(yield func(*player) (next bool)) {
	pls.mu.RLock()
	players := make([]*player, 0, len(pls.items))
	for pl := range pls.items {
		players = append(players, pl)
	}
	pls.mu.RUnlock()

	for _, item := range players {
		if !yield(item) {
			break
		}
	}
}

func (pls *players) WaitAndPoll(
	ctx context.Context,
	onUpdate func(update playerUpdate),
	onEvent func(event playerEvent),
	onFinish func(*player),
) error {
	pls.mu.Lock()
	if pls.waitErrGroup != nil {
		return errors.New("already waiting")
	}

	pls.onUpdate = onUpdate
	pls.onEvent = onEvent
	pls.onFinish = onFinish
	pls.waitErrGroup, pls.waitCtx = errgroup.WithContext(ctx)

	for pl := range pls.items {
		pls.startWaitingForPlayer(pl)
	}
	pls.mu.Unlock()

	err := pls.waitErrGroup.Wait()

	pls.mu.Lock()
	pls.waitCtx = nil
	pls.waitErrGroup = nil
	pls.mu.Unlock()

	return err
}

func (pls *players) startWaitingForPlayer(pl *player) {
	pls.waitErrGroup.Go(func() error {
		err := pl.WaitAndPoll(pls.waitCtx, pls.onUpdate, pls.onEvent, pls.onFinish)
		pls.mu.Lock()
		defer pls.mu.Unlock()
		delete(pls.items, pl)
		if !errors.Is(err, instance.ErrInstanceFinished) {
			return err
		}
		if len(pls.items) == 0 {
			return ErrAllInstancesFinished
		}
		return nil
	})
}
