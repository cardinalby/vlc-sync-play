package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/cardinalby/vlc-sync-play/internal/app"
	"github.com/cardinalby/vlc-sync-play/internal/cli/args"
	"github.com/cardinalby/vlc-sync-play/internal/cli/debug"
	"github.com/cardinalby/vlc-sync-play/internal/cli/interactive"
	"github.com/cardinalby/vlc-sync-play/pkg/util/logging"
)

type cliApp interface {
	Init(settingsPatch app.SettingsPatch) error
	Start(ctx context.Context) error
}

func RunCliApp(ctx context.Context) error {
	cmdLineArgs, err := args.ParseCmdLineArgs()
	if err != nil {
		return fmt.Errorf("error parsing command line args: %w", err)
	}

	var cliApp cliApp
	if cmdLineArgs.Debug != nil && *cmdLineArgs.Debug {
		cliApp = debug.NewApp(logging.NewLogger(os.Stdout))
	} else {
		cliApp = interactive.NewApp(logging.NewNopLogger())
	}
	if err := cliApp.Init(cmdLineArgs); err != nil {
		if errors.Is(err, app.ErrSettingsPatchIsInvalid) {
			return fmt.Errorf("invalid command line args: %w", err)
		}
		return fmt.Errorf("error initializing cli app: %w", err)
	}
	return cliApp.Start(ctx)
}
