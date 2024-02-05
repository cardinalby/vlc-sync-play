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
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/extended/repetition"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/timings"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/instance"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/state"
	"golang.org/x/sync/errgroup"
)

var ErrAllInstancesFinished = errors.New("all instances finished")

type playerUpdate struct {
	player *player
	update state.Update
}

type playerEvent struct {
	event  instance.StdErrEvent
	player *player
}

func (pu *playerUpdate) GetSyncCommands() extended.CmdGroup {
	return pu.player.client.state.GetSyncCommands(pu.update.ChangedProps)
}

type playerSettings struct {
	pollingInterval typeutil.Observable[time.Duration]
	stdErrEvents    typeutil.Observable[instance.EventsToParse]
}

type player struct {
	instance *instance.Instance
	client   *PollingClient
	settings playerSettings
}

func newPlayer(
	instance *instance.Instance,
	settings playerSettings,
	parentLogger logging.Logger,
) *player {
	return &player{
		instance: instance,
		client: newClient(
			instance.Client,
			settings.pollingInterval,
			parentLogger.WithPrefix(fmt.Sprintf("P[%d]", instance.ID)),
		),
		settings: settings,
	}
}

func (pl *player) WaitAndPoll(
	ctx context.Context,
	onUpdate func(update playerUpdate),
	onEvent func(event playerEvent),
	onFinish func(*player),
) error {
	pl.instance.SetEventsToParse(pl.settings.stdErrEvents.GetValue())
	defer pl.settings.stdErrEvents.Subscribe(func(events instance.EventsToParse) {
		pl.instance.SetEventsToParse(events)
	}).Unsubscribe()

	errGr, ctx := errgroup.WithContext(ctx)

	errGr.Go(func() error {
		err := pl.instance.Wait(ctx, func(event instance.StdErrEvent) {
			onEvent(playerEvent{event: event, player: pl})
		})
		onFinish(pl)
		return err
	})
	errGr.Go(func() error {
		return pl.client.StartPolling(ctx, func(stateUpdate state.Update) {
			pl.onUpdate(stateUpdate, onUpdate)
		})
	})

	return errGr.Wait()
}

func (pl *player) SendCmdGroup(
	ctx context.Context,
	cmdGroup extended.CmdGroup,
	rule repetition.Rule,
) (statusEx *basic.StatusEx, err error) {
	return pl.client.SendCmdGroup(ctx, cmdGroup, rule)
}

func (pl *player) onUpdate(stateUpdate state.Update, notify func(playerUpdate)) {
	if stateUpdate.ChangedProps.HasState() && stateUpdate.Status.State == basic.PlaybackStateStopped {
		// "stopped" state can be caused by player instance shutdown (reproduces mainly on Windows).
		// If player is not shut down soon, send update as normal, skip update otherwise
		// to avoid stopping all players.
		timer := time.NewTimer(timings.WaitForShutdownAfterStopDuration)
		defer timer.Stop()
		select {
		case <-timer.C:
		case <-pl.instance.Finished():
			return
		}
	}
	notify(playerUpdate{
		player: pl,
		update: stateUpdate,
	})
}

func (pl *player) IsRecoverableErr(err error) bool {
	return pl.client.IsRecoverableErr(err)
}

func (pl *player) GetID() uint {
	return pl.instance.ID
}
