package instance

import (
	"context"
	"errors"
	"os/exec"

	"github.com/cardinalby/vlc-sync-play/pkg/util/logging"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/extended"
	"golang.org/x/sync/errgroup"
)

type Instance struct {
	Client       *extended.Client
	Cmd          *exec.Cmd
	stdErrParser *OutputParser
	logger       logging.Logger
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
	ctx, cancel := context.WithCancel(ctx)
	errGr, ctx := errgroup.WithContext(ctx)

	errGr.Go(func() error {
		return i.stdErrParser.Start(ctx, onStderrEvent)
	})
	errGr.Go(func() error {
		defer cancel()
		return i.Cmd.Wait()
	})
	errGr.Go(func() error {
		<-ctx.Done()
		_ = i.Stop()
		return ctx.Err()
	})

	err := errGr.Wait()
	if errors.Is(err, ctx.Err()) {
		return nil
	}
	return err
}
