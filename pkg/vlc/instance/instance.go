package instance

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic"
)

type Instance struct {
	ApiClient basic.ApiClient
	Cmd       *exec.Cmd
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

func (i Instance) Wait(ctx context.Context) error {
	waitRes := make(chan error, 1)
	go func() {
		waitRes <- i.Cmd.Wait()
		close(waitRes)
	}()

	select {
	case <-ctx.Done():
		fmt.Printf("context done, killing process %d\n", i.GetPID())
		_ = i.Stop()
		return ctx.Err()
	case err := <-waitRes:
		return err
	}
}
