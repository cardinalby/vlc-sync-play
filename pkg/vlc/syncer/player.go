package syncer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cardinalby/vlc-sync-play/pkg/util/logging"
	typeutil "github.com/cardinalby/vlc-sync-play/pkg/util/rx"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/extended"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/instance"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/state"
)

var ErrInstanceFinished = errors.New("instance finished")
var ErrInstanceFailed = errors.New("instance failed")

type playerSettings struct {
	pollingInterval typeutil.Observable[time.Duration]
	stdErrEvents    typeutil.Observable[instance.EventsToParse]
}

type player struct {
	id       uint
	instance instance.Instance
	client   *Client
	settings playerSettings
}

type playerUpdate struct {
	player *player
	update state.Update
	status basic.StatusEx
}

type playerEvent struct {
	event  instance.StdErrEvent
	player *player
}

func (pu *playerUpdate) GetSyncCommands() extended.CmdGroup {
	return pu.player.client.state.GetSyncCommands(pu.update.ChangedProps)
}

func newPlayer(
	instance instance.Instance,
	settings playerSettings,
	logger logging.Logger,
	id uint,
) *player {
	return &player{
		instance: instance,
		client: newClient(
			instance.Client,
			settings.pollingInterval,
			logger.WithPrefix(fmt.Sprintf("P[%d]", id)),
		),
		settings: settings,
	}
}

func (pl *player) StartWaiting(
	ctx context.Context,
	goroutineRunner func(func() error),
	onUpdate func(update playerUpdate),
	onStderrEvent func(event instance.StdErrEvent),
) {
	pl.instance.SetEventsToParse(pl.settings.stdErrEvents.GetValue())
	stdErrEventsObserver := pl.settings.stdErrEvents.Subscribe(func(events instance.EventsToParse) {
		pl.instance.SetEventsToParse(events)
	})

	goroutineRunner(func() error {
		err := pl.instance.Wait(ctx, onStderrEvent)
		stdErrEventsObserver.Unsubscribe()
		if err != nil {
			return fmt.Errorf("%w: %s", ErrInstanceFailed, err.Error())
		}
		return ErrInstanceFinished
	})
	goroutineRunner(func() error {
		return pl.client.StartPolling(ctx, func(update state.Update) {
			onUpdate(playerUpdate{
				player: pl,
				update: update,
			})
		})
	})
}

func (pl *player) SendCmdGroup(
	ctx context.Context,
	cmdGroup extended.CmdGroup,
) (statusEx *basic.StatusEx, shouldRepeat bool) {
	statusEx, shouldRepeat = pl.client.SendCmdGroup(ctx, cmdGroup)
	return statusEx, shouldRepeat && pl.instance.IsRunning()
}
