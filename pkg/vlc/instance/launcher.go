package instance

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cardinalby/vlc-sync-play/internal/app/static_features"
	"github.com/cardinalby/vlc-sync-play/pkg/util/logging"
	typeutil "github.com/cardinalby/vlc-sync-play/pkg/util/type"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic/protocols"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/extended"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/extended/repetition"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/timings"
)

const IDNone = 0

type LaunchOptions struct {
	FileURI typeutil.Optional[string]
	NoVideo bool
}

type Launcher interface {
	Launch(ctx context.Context, options LaunchOptions) (*Instance, error)
}

func NewLauncher(
	vlcPath string,
	apiProtocol protocols.ApiProtocol,
	logger logging.Logger,
) Launcher {
	return &launcher{
		nextId:      IDNone + 1,
		vlcPath:     vlcPath,
		apiProtocol: apiProtocol,
		logger:      logger,
	}
}

type launcher struct {
	nextId      uint
	vlcPath     string
	apiProtocol protocols.ApiProtocol
	logger      logging.Logger
}

func (l *launcher) Launch(ctx context.Context, options LaunchOptions) (inst *Instance, err error) {
	apiClient, err := protocols.NewLocalBasicApiClient(l.apiProtocol, l.logger)
	if err != nil {
		return nil, err
	}
	args := l.getArgs(apiClient, options)

	cmd, outputParser, err := l.startCmd(args)
	if err != nil {
		return nil, err
	}

	inst = newInstance(
		l.nextId,
		apiClient,
		cmd,
		outputParser,
		l.logger,
	)
	l.nextId++

	if err := l.waitUntilReady(ctx, inst, options.FileURI.Value); err != nil {
		return nil, err
	}

	return inst, err
}

func (l *launcher) getArgs(apiClient basic.ApiClient, options LaunchOptions) []string {
	args := apiClient.GetLaunchArgs()
	if options.NoVideo {
		args = append(args, "--no-video")
	}
	if static_features.ClickPause {
		args = append(args, "--verbose", "2")
	}
	if static_features.LaunchWithFile && options.FileURI.HasValue {
		args = append(args, options.FileURI.Value)
	}
	return args
}

func (l *launcher) startCmd(args []string) (*exec.Cmd, *OutputParser, error) {
	l.logger.Info("Launching %s %s", l.vlcPath, strings.Join(args, " "))
	cmd := exec.Command(l.vlcPath, args...)
	if workingDir, err := os.Getwd(); err == nil {
		cmd.Dir = workingDir
	}
	cmd.Env = os.Environ()

	var outputParser *OutputParser
	if static_features.ClickPause {
		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			return nil, nil, fmt.Errorf("error creating stderr pipe: %w", err)
		}
		outputParser = NewOutputParser(stderrPipe, EventsToParse{})
	} else {
		outputParser = NewOutputParser(nil, EventsToParse{})
	}
	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("error starting vlc process: %w", err)
	}
	return cmd, outputParser, nil
}

func (l *launcher) waitUntilReady(
	ctx context.Context,
	inst *Instance,
	fileURI string,
) error {
	rule := repetition.WithInterval(timings.WaitUntilOnlinePollingInterval)

	l.logger.Info("* waiting until online")
	if _, err := inst.Client.GetStatusEx(ctx, rule); err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}
	l.logger.Info("* online")

	//goland:noinspection GoBoolExpressions
	if fileURI != "" && !static_features.LaunchWithFile {
		l.logger.Info("* waiting open file cmd to complete")
		if _, err := inst.Client.SendCmdGroup(
			ctx,
			extended.CmdGroup{
				OpenFile: typeutil.NewOptional(fileURI),
			},
			rule,
		); err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		l.logger.Info("* file opened")
	}
	return nil
}
