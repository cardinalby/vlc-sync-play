package args

import (
	"flag"
	"os"
	"time"

	"github.com/cardinalby/vlc-sync-play/internal/app"
	"github.com/cardinalby/vlc-sync-play/pkg/util/cliargs"
	"golang.org/x/exp/slices"
)

type CmdLineArgs struct {
	VlcPath           *string  `flag:"vlc" usage:"Path to VLC executable"`
	InstancesNumber   *int     `flag:"instances" usage:"Number of VLC instances"`
	PollingIntervalMs *int64   `flag:"interval" usage:"Polling interval ms"`
	ClickPause        *bool    `flag:"click-pause" usage:"Click to pause/resume playback"`
	NoVideo           *bool    `flag:"no-video" usage:"Start additional instances without video"`
	ReSeekSrc         *bool    `flag:"re-seek-src" usage:"Re-seek source player for precise sync"`
	Debug             *bool    `flag:"debug" usage:"Debug mode"`
	FilePaths         []string `flagArgs:"true"`
}

func (args CmdLineArgs) ApplyToSettings(s *app.Settings) (updated bool) {
	if args.VlcPath != nil {
		s.VlcPath = *args.VlcPath
		updated = true
	}
	if args.InstancesNumber != nil {
		s.InstancesNumber.SetValue(*args.InstancesNumber)
		updated = true
	}
	if args.PollingIntervalMs != nil {
		s.PollingInterval.SetValue(time.Duration(*args.PollingIntervalMs) * time.Millisecond)
		updated = true
	}
	if args.ClickPause != nil {
		s.ClickPause.SetValue(*args.ClickPause)
		updated = true
	}
	if args.NoVideo != nil {
		s.NoVideo.SetValue(*args.NoVideo)
		updated = true
	}
	if args.ReSeekSrc != nil {
		s.ReSeekSrc.SetValue(*args.ReSeekSrc)
		updated = true
	}
	if !slices.Equal(s.FilePaths, args.FilePaths) {
		s.FilePaths = args.FilePaths
		updated = true
	}
	return updated
}

func ParseCmdLineArgs() (a CmdLineArgs, err error) {
	var args CmdLineArgs
	return args, cliargs.ParseStruct(&args, "", flag.ExitOnError, os.Args[1:])
}
