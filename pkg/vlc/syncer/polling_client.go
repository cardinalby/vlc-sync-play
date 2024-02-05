package syncer

import (
	"context"
	"time"

	"github.com/cardinalby/vlc-sync-play/pkg/util/logging"
	typeutil "github.com/cardinalby/vlc-sync-play/pkg/util/rx"
	timeutil "github.com/cardinalby/vlc-sync-play/pkg/util/time"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/extended"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/extended/repetition"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/timings"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/state"
)

type PollingClient struct {
	client          *extended.Client
	pollingInterval typeutil.Observable[time.Duration]
	state           *state.State
	logger          logging.Logger
}

func newClient(
	client *extended.Client,
	pollingInterval typeutil.Observable[time.Duration],
	logger logging.Logger,
) *PollingClient {
	return &PollingClient{
		client:          client,
		pollingInterval: pollingInterval,
		state:           state.NewState(logger),
		logger:          logger,
	}
}

func (c *PollingClient) StartPolling(ctx context.Context, onUpdate func(state.Update)) error {
	for {
		newStatus, err := c.client.GetStatusEx(ctx, repetition.Single())
		if err == nil {
			if err := c.onNewStatus(ctx, newStatus, onUpdate); err != nil {
				return err
			}
		} else if !c.IsRecoverableErr(err) {
			// Not recoverable
			return err
		}
		if err := timeutil.SleepCtx(ctx, c.pollingInterval.GetValue()); err != nil {
			// Not recoverable
			return err
		}
	}
}

func (c *PollingClient) SendCmdGroup(
	ctx context.Context,
	group extended.CmdGroup,
	rule repetition.Rule,
) (statusEx *basic.StatusEx, err error) {
	c.logger.Info("SendCmdGroup")
	res, err := c.client.SendCmdGroup(ctx, group, rule)
	if res != nil {
		c.logger.Info("apply Cmd status: %s", res.String())
		c.state.ApplyNewStatus(res)
	}
	return res, err
}

func (c *PollingClient) IsRecoverableErr(err error) bool {
	return c.client.IsRecoverableErr(err)
}

func (c *PollingClient) onNewStatus(
	ctx context.Context,
	newStatus basic.StatusEx,
	onUpdate func(state.Update),
) error {
	oldStateStr := c.state.String()
	update, err := c.state.GetUpdate(&newStatus)
	c.state.ApplyNewStatus(&newStatus)
	if err != nil || !update.ChangedProps.HasAny() || update.IsNatural {
		// ignore errors or no changes
		return nil
	}

	c.logger.Info("Update: %s. \nOld: %s, \nNew: %s", &update, oldStateStr, c.state)
	onUpdate(update)

	if update.ChangedProps.HasFileURI() {
		durationSinceFileOpened := time.Since(newStatus.Moment.Center())
		// Wait for auto-seek after file opened
		if err := timeutil.SleepCtx(
			ctx,
			timings.WaitForAutoSeekAfterFileOpenedDuration-durationSinceFileOpened,
		); err != nil {
			return err
		}
	}

	return nil
}
