package extended

import (
	"context"
	"sync"
	"time"

	urlutil "github.com/cardinalby/vlc-sync-play/pkg/url"
	"github.com/cardinalby/vlc-sync-play/pkg/util/logging"
	mathutil "github.com/cardinalby/vlc-sync-play/pkg/util/math"
	timeutil "github.com/cardinalby/vlc-sync-play/pkg/util/time"
	typeutil "github.com/cardinalby/vlc-sync-play/pkg/util/type"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/extended/repetition"
	"golang.org/x/sync/errgroup"
)

const respTimeSamplesCount = 10

type lastStatusPart struct {
	Filename  string
	LengthSec int
	FileURI   string
}

type Client struct {
	api                      basic.ApiClient
	mu                       sync.Mutex
	statusRespTime           *mathutil.AvgAcc[time.Duration]
	lastStatusPart           typeutil.Optional[lastStatusPart]
	getInstanceFinishedError func() error
	isInstanceFinishedError  func(error) bool
	logger                   logging.Logger
}

func NewClient(
	api basic.ApiClient,
	getInstanceFinishedError func() error,
	isInstanceFinishedError func(error) bool,
	logger logging.Logger,
) *Client {
	return &Client{
		api:                      api,
		statusRespTime:           mathutil.NewAvgAcc[time.Duration](respTimeSamplesCount),
		getInstanceFinishedError: getInstanceFinishedError,
		isInstanceFinishedError:  isInstanceFinishedError,
		logger:                   logger,
	}
}

func (c *Client) GetStatusEx(ctx context.Context, rule repetition.Rule) (basic.StatusEx, error) {
	status, err := c.getStatus(ctx, rule)

	if err == nil {
		return c.addFileURIToStatus(ctx, status, rule)
	}
	return basic.StatusEx{}, err
}

func (c *Client) SendCmdGroup(
	ctx context.Context,
	group CmdGroup,
	rule repetition.Rule,
) (res *basic.StatusEx, err error) {
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

		statusEx, err := c.sendStatusCmd(ctx, cmd, rule)
		if err := updateRes(statusEx, err); err != nil {
			return nil, err
		}
		clarificationStartedAt := time.Now()
		// Ensure that file is opened. It can be not reported as opened in the first request.
		// Also, VLC may go crazy if you send next commands immediately
		for !urlutil.EqualIgnoreSchema(statusEx.FileURI, targetFileURI) {
			statusEx, err = c.GetStatusEx(ctx, repetition.Single())
			if err := updateRes(statusEx, err); err != nil {
				return nil, err
			}
		}
		c.logger.Info("File opened in %s", time.Since(clarificationStartedAt).String())
	}

	errGr, ctx := errgroup.WithContext(ctx)

	if isNotStopped && group.Seek.HasValue {
		errGr.Go(func() error {
			return updateRes(c.sendStatusCmd(ctx, group.GetSeekCmd(c.getCmdExpectedExecutionTime()), rule))
		})
	}

	if isNotStopped && group.Rate.HasValue {
		errGr.Go(func() error {
			return updateRes(c.sendStatusCmd(ctx, group.GetRateCmd(), rule))
		})
	}

	if cmd := group.GetStateCmd(); cmd != nil {
		errGr.Go(func() error {
			return updateRes(c.sendStatusCmd(ctx, cmd, rule))
		})
	}

	err = errGr.Wait()
	return res, err
}

func (c *Client) getCmdExpectedExecutionTime() time.Time {
	if avg, ok := c.statusRespTime.Avg(); ok {
		return time.Now().Add(avg / 2)
	}
	// should not happen
	return time.Now()
}

func (c *Client) addFileURIToStatus(
	ctx context.Context,
	status basic.Status,
	rule repetition.Rule,
) (basic.StatusEx, error) {
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
			if fileURI, err = c.getCurrentFileUri(ctx, rule); err != nil {
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
	return c.api.IsRecoverableErr(err) && !c.isInstanceFinishedError(err)
}

func (c *Client) getStatus(
	ctx context.Context,
	rule repetition.Rule,
) (status basic.Status, err error) {
	err = c.apiCall(ctx, func() error {
		status, err = c.api.GetStatus(ctx)
		return err
	}, rule)
	return status, err
}

func (c *Client) getCurrentFileUri(ctx context.Context, rule repetition.Rule) (fileURI string, err error) {
	err = c.apiCall(ctx, func() error {
		fileURI, err = c.api.GetCurrentFileUri(ctx)
		return err
	}, rule)
	return fileURI, err
}

func (c *Client) sendStatusCmd(
	ctx context.Context,
	cmd basic.Command,
	rule repetition.Rule,
) (statusEx basic.StatusEx, err error) {
	var status basic.Status
	if err = c.apiCall(ctx, func() error {
		status, err = c.api.SendStatusCmd(ctx, cmd)
		return err
	}, rule); err != nil {
		return statusEx, err
	}
	return c.addFileURIToStatus(ctx, status, rule)
}

func (c *Client) apiCall(ctx context.Context, apiAction func() error, rule repetition.Rule) error {
	for {
		err := apiAction()
		if err == nil {
			return nil
		}
		if !c.IsRecoverableErr(err) || !rule.Interval.HasValue {
			return err
		}
		if err := timeutil.SleepCtx(ctx, rule.Interval.Value); err != nil {
			return err
		}
	}
}
