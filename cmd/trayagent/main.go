package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/cardinalby/vlc-sync-play/internal/app"
	"github.com/cardinalby/vlc-sync-play/internal/args"
	"github.com/cardinalby/vlc-sync-play/internal/trayagent"
	"github.com/cardinalby/vlc-sync-play/pkg/util/logging"
)

func main() {
	if err := runTrayAgentApp(); err != nil {
		fmt.Printf(err.Error())
	}
}

func runTrayAgentApp() error {
	cmdLineArgs, err := args.ParseCmdLineArgs()
	if err != nil {
		return fmt.Errorf("error parsing command line args: %w", err)
	}
	var logger logging.Logger
	if cmdLineArgs.Debug {
		logger = logging.NewLogger(os.Stdout)
	} else {
		logger = logging.NewNopLogger()
	}

	taApp := trayagent.NewApp(logger)
	if err := taApp.Init(cmdLineArgs); err != nil {
		if errors.Is(err, app.ErrSettingsPatchIsInvalid) {
			return fmt.Errorf("invalid command line args: %w", err)
		}
		return fmt.Errorf("error initializing app: %w", err)
	}
	return taApp.Start(context.Background())
}
