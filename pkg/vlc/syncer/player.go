package syncer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/extended"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/instance"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/state"
)

var ErrInstanceFinished = errors.New("instance finished")
var ErrInstanceFailed = errors.New("instance failed")

type player struct {
	instance instance.Instance
	client   *Client
}

type playerUpdate struct {
	player *player
	update state.Update
	status basic.StatusEx
}

func (pu *playerUpdate) GetSyncCommands() extended.CmdGroup {
	return pu.player.client.state.GetSyncCommands(pu.update.ChangedProps)
}

func newPlayer(instance instance.Instance, pollingInterval time.Duration) *player {
	return &player{
		instance: instance,
		client: newClient(
			extended.NewClient(instance.ApiClient),
			pollingInterval,
		),
	}
}

func (pl *player) StartWaiting(
	ctx context.Context,
	goroutineRunner func(func() error),
	updatesChan chan<- playerUpdate,
) {
	goroutineRunner(func() error {
		err := pl.instance.Wait(ctx)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrInstanceFailed, err.Error())
		}
		return ErrInstanceFinished
	})
	goroutineRunner(func() error {
		return pl.client.StartPolling(ctx)
	})
	goroutineRunner(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case stateUpdate, ok := <-pl.client.updatesChan:
				if !ok {
					return ctx.Err()
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case updatesChan <- playerUpdate{
					player: pl,
					update: stateUpdate,
				}:
				}
			}
		}
	})
}

func (pl *player) SendCmdGroup(
	ctx context.Context,
	cmdGroup extended.CmdGroup,
) (statusEx *basic.StatusEx, shouldRepeat bool) {
	statusEx, shouldRepeat = pl.client.SendCmdGroup(ctx, cmdGroup)
	return statusEx, shouldRepeat && pl.instance.IsRunning()
}
