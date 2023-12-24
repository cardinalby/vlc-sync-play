package syncer

import (
	"context"
	"time"

	"github.com/cardinalby/vlc-sync-play/pkg/util/logging"
	typeutil "github.com/cardinalby/vlc-sync-play/pkg/util/rx"
	timeutil "github.com/cardinalby/vlc-sync-play/pkg/util/time"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/extended"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/timings"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/state"
)

type Client struct {
	client          *extended.Client
	pollingInterval typeutil.Observable[time.Duration]
	state           *state.State
	logger          logging.Logger
}

func newClient(
	client *extended.Client,
	pollingInterval typeutil.Observable[time.Duration],
	logger logging.Logger,
) *Client {
	return &Client{
		client:          client,
		pollingInterval: pollingInterval,
		state:           state.NewState(logger),
		logger:          logger,
	}
}

func (c *Client) StartPolling(ctx context.Context, onUpdate func(state.Update)) error {
	for {
		newStatus, err := c.client.GetStatusEx(ctx)
		if err == nil {
			if err := c.onNewStatus(ctx, newStatus, onUpdate); err != nil {
				return err
			}
		} else if !c.client.IsRecoverableErr(err) {
			// Not recoverable
			return err
		}
		if err := timeutil.SleepCtx(ctx, c.pollingInterval.GetValue()); err != nil {
			// Not recoverable
			return err
		}
	}
}

func (c *Client) SendCmdGroup(
	ctx context.Context,
	group extended.CmdGroup,
) (statusEx *basic.StatusEx, shouldRepeat bool) {
	c.logger.Info("SendCmdGroup")
	res, err := c.client.SendCmdGroup(ctx, group)
	if res != nil {
		c.state.ApplyNewStatus(res)
	}
	return res, err != nil && c.client.IsRecoverableErr(err)
}

func (c *Client) onNewStatus(
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
		if err := timeutil.SleepCtx(ctx, timings.WaitForAutoSeekAfterFileOpenedDuration-durationSinceFileOpened); err != nil {
			return err
		}
	}

	return nil
}
