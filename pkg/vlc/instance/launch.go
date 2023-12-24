package instance

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/cardinalby/vlc-sync-play/pkg/util/logging"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic/protocols"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/extended"
)

type LaunchOptions struct {
	FilePaths []string
	NoVideo   bool
}

type Launcher func(ctx context.Context, options LaunchOptions) (Instance, error)

func GetLauncher(
	vlcPath string,
	apiProtocol protocols.ApiProtocol,
	logger logging.Logger,
) Launcher {
	return func(ctx context.Context, options LaunchOptions) (Instance, error) {
		apiClient, err := protocols.NewLocalBasicApiClient(apiProtocol, logger)
		if err != nil {
			return Instance{}, err
		}

		args := apiClient.GetLaunchArgs()
		if options.NoVideo {
			args = append(args, "--no-video")
		}
		args = append(args, options.FilePaths...)
		args = append(args, "--verbose", "2")
		cmd := exec.Command(vlcPath, args...)
		if workingDir, err := os.Getwd(); err == nil {
			cmd.Dir = workingDir
		}
		cmd.Env = os.Environ()

		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			return Instance{}, fmt.Errorf("error creating stderr pipe: %w", err)
		}
		if err := cmd.Start(); err != nil {
			return Instance{}, fmt.Errorf("error starting vlc process: %w", err)
		}

		if client, err := extended.CreateClientWaitOnline(ctx, apiClient, logger); err == nil {
			return Instance{
				Client:       client,
				Cmd:          cmd,
				stdErrParser: NewOutputParser(stderrPipe, EventsToParse{}),
				logger:       logger,
			}, nil
		} else {
			return Instance{}, err
		}
	}
}
