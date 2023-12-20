package syncer

import (
	"context"
	"fmt"
	"time"

	timeutil "github.com/cardinalby/vlc-sync-play/pkg/util/time"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/extended"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/state"
)

var index int

type Client struct {
	// todo: remove index
	index           int
	client          *extended.Client
	pollingInterval time.Duration
	state           *state.State
	updatesChan     chan state.Update
}

func newClient(
	client *extended.Client,
	pollingInterval time.Duration,
) *Client {
	index++
	return &Client{
		index:           index,
		client:          client,
		pollingInterval: pollingInterval,
		state:           state.NewState(),
		updatesChan:     make(chan state.Update),
	}
}

func (c *Client) StartPolling(ctx context.Context) error {
	defer func() {
		close(c.updatesChan)
	}()

	for {
		newStatus, err := c.client.GetStatusEx(ctx)
		if err == nil {
			_ = c.onNewStatus(ctx, newStatus)
		} else if !c.client.IsRecoverableErr(err) {
			// Not recoverable
			return err
		}
		if err := timeutil.SleepCtx(ctx, c.pollingInterval); err != nil {
			// Not recoverable
			return err
		}
	}
}

func (c *Client) SendCmdGroup(
	ctx context.Context,
	group extended.CmdGroup,
) (statusEx *basic.StatusEx, shouldRepeat bool) {
	fmt.Printf("Client %d: SendCmdGroup\n", c.index)
	res, err := c.client.SendCmdGroup(ctx, group)
	if res != nil {
		c.state.ApplyNewStatus(res)
	}
	return res, err != nil && c.client.IsRecoverableErr(err)
}

func (c *Client) onNewStatus(ctx context.Context, newStatus basic.StatusEx) (err error) {
	update, err := c.state.ApplyNewStatusAndGetUpdate(&newStatus)
	if err != nil || !update.ChangedProps.HasAny() {
		// ignore errors or no changes
		return nil
	}
	select {
	case c.updatesChan <- update:
	case <-ctx.Done():
	}
	return ctx.Err()
}
