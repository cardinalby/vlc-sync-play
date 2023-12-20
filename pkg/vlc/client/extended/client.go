package extended

import (
	"context"
	"fmt"
	"sync"
	"time"

	mathutil "github.com/cardinalby/vlc-sync-play/pkg/util/math"
	typeutil "github.com/cardinalby/vlc-sync-play/pkg/util/type"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic"
	"golang.org/x/sync/errgroup"
)

const respTimeSamplesCount = 10
const respTimeDefaultValue = 20 * time.Millisecond

type lastStatusPart struct {
	Filename  string
	LengthSec int
	FileURI   string
}

type Client struct {
	api            basic.ApiClient
	mu             sync.Mutex
	statusRespTime *mathutil.AvgAcc[time.Duration]
	lastStatusPart typeutil.Optional[lastStatusPart]
}

func NewClient(api basic.ApiClient) *Client {
	return &Client{
		api:            api,
		statusRespTime: mathutil.NewAvgAcc[time.Duration](respTimeSamplesCount, respTimeDefaultValue),
	}
}

func (c *Client) GetStatusEx(ctx context.Context) (basic.StatusEx, error) {
	status, err := c.api.GetStatus(ctx)

	if err == nil {
		return c.getStatusEx(ctx, status)
	}
	return basic.StatusEx{}, err
}

func (c *Client) SendStatusCmd(ctx context.Context, cmd basic.Command) (basic.StatusEx, error) {
	status, err := c.api.SendStatusCmd(ctx, cmd)
	if err == nil {
		res, err := c.getStatusEx(ctx, status)
		return res, err
	}
	return basic.StatusEx{}, err
}

func (c *Client) SendCmdGroup(ctx context.Context, group CmdGroup) (res *basic.StatusEx, err error) {
	resMu := sync.Mutex{}
	updateRes := func(statusEx basic.StatusEx, err error) error {
		if err == nil {
			resMu.Lock()
			res = &statusEx
			resMu.Unlock()
		}
		return err
	}

	isNotStopped := !group.State.HasValue || group.State.Value != basic.PlaybackStateStopped

	if group.OpenFile.HasValue && isNotStopped {
		targetFileURI := group.OpenFile.Value
		cmd := group.GetOpenFileCmd()

		statusEx, err := c.SendStatusCmd(ctx, cmd)
		if err := updateRes(statusEx, err); err != nil {
			return nil, err
		}

		// todo: trying to cope with vlc no progress showing
		for statusEx.FileURI != targetFileURI {
			fmt.Printf("=== Client: waiting for file change\n")
			time.Sleep(time.Second)
			statusEx, err = c.GetStatusEx(ctx)
			if err != nil {
				return nil, err
			}
			time.Sleep(time.Second)
		}
	}

	errGr, ctx := errgroup.WithContext(ctx)

	if isNotStopped && group.Seek.HasValue {
		errGr.Go(func() error {
			return updateRes(c.SendStatusCmd(ctx, group.GetSeekCmd(c.getCmdExpectedExecutionTime())))
		})
	}

	if isNotStopped && group.Rate.HasValue {
		errGr.Go(func() error {
			return updateRes(c.SendStatusCmd(ctx, group.GetRateCmd()))
		})
	}

	if cmd := group.GetStateCmd(); cmd != nil {
		errGr.Go(func() error {
			return updateRes(c.SendStatusCmd(ctx, cmd))
		})
	}

	err = errGr.Wait()
	return res, err
}

func (c *Client) getCmdExpectedExecutionTime() time.Time {
	return time.Now().Add(c.statusRespTime.Avg() / 2)
}

func (c *Client) getStatusEx(ctx context.Context, status basic.Status) (basic.StatusEx, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.statusRespTime.Add(status.Moment.Length())

	fileURI := c.lastStatusPart.Value.FileURI

	if !c.lastStatusPart.HasValue || (status.FileName != c.lastStatusPart.Value.Filename ||
		status.LengthSec != c.lastStatusPart.Value.LengthSec) {
		// file changed or playback stopped
		if status.State == basic.PlaybackStateStopped {
			fileURI = ""
		} else {
			var err error
			if fileURI, err = c.api.GetCurrentFileUri(ctx); err != nil {
				return basic.StatusEx{}, err
			}
		}
	}

	c.lastStatusPart.Set(lastStatusPart{
		Filename:  status.FileName,
		LengthSec: status.LengthSec,
		FileURI:   fileURI,
	})

	return basic.StatusEx{
		Status:  status,
		FileURI: fileURI,
	}, nil
}

func (c *Client) IsRecoverableErr(err error) bool {
	return c.api.IsRecoverableErr(err)
}
