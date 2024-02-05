package instance

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/cardinalby/vlc-sync-play/pkg/util/logging"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/extended"
	"golang.org/x/sync/errgroup"
)

var ErrInstanceFinished = errors.New("instance finished")
var ErrInstanceFailed = fmt.Errorf("%w with error", ErrInstanceFinished)

type Instance struct {
	ID              uint
	Client          *extended.Client
	Cmd             *exec.Cmd
	stdErrParser    *OutputParser
	logger          logging.Logger
	internalWaitErr chan error
}

func newInstance(
	id uint,
	api basic.ApiClient,
	cmd *exec.Cmd,
	stdErrParser *OutputParser,
	logger logging.Logger,
) *Instance {
	inst := &Instance{
		ID:              id,
		Cmd:             cmd,
		stdErrParser:    stdErrParser,
		logger:          logger,
		internalWaitErr: make(chan error, 1),
	}
	inst.Client = extended.NewClient(
		api,
		inst.GetFinishedError,
		func(err error) bool {
			return errors.Is(err, ErrInstanceFinished)
		},
		logger,
	)
	inst.startInternalWait()
	return inst
}

func (i Instance) GetPID() int {
	if i.Cmd.Process == nil {
		return 0
	}
	return i.Cmd.Process.Pid
}

func (i Instance) IsRunning() bool {
	return i.Cmd.Process != nil
}

func (i Instance) GetFinishedError() error {
	if !i.IsRunning() {
		return ErrInstanceFinished
	}
	return nil
}

func (i Instance) Stop() error {
	if i.IsRunning() {
		return i.Cmd.Process.Kill()
	}
	return nil
}

func (i Instance) SetEventsToParse(events EventsToParse) {
	i.stdErrParser.SetEventsToParse(events)
}

func (i Instance) Wait(ctx context.Context, onStderrEvent func(event StdErrEvent)) error {
	errGr, ctx := errgroup.WithContext(ctx)

	errGr.Go(func() error {
		return i.stdErrParser.Start(ctx, onStderrEvent)
	})
	errGr.Go(func() error {
		return <-i.internalWaitErr
	})
	errGr.Go(func() error {
		<-ctx.Done()
		_ = i.Stop()
		return nil
	})

	return errGr.Wait()
}

func (i Instance) Finished() <-chan error {
	return i.internalWaitErr
}

func (i Instance) startInternalWait() {
	go func() {
		err := i.Cmd.Wait()
		if err == nil {
			err = ErrInstanceFinished
		} else {
			err = errors.Join(ErrInstanceFailed, err)
		}
		i.internalWaitErr <- err
		close(i.internalWaitErr)
	}()
}
