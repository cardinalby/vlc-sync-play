package args

import (
	"flag"
	"os"
	"time"

	flago "github.com/cardinalby/go-struct-flags"
	"github.com/cardinalby/vlc-sync-play/internal/app/static_features"

	"github.com/cardinalby/vlc-sync-play/internal/app"
	"golang.org/x/exp/slices"
)

type CmdLineArgs struct {
	VlcPath           *string  `flag:"vlc" flagUsage:"Path to VLC executable"`
	InstancesNumber   *int     `flag:"instances" flagUsage:"Number of VLC instances"`
	PollingIntervalMs *int64   `flag:"interval" flagUsage:"Polling interval ms"`
	ClickPause        *bool    `flag:"click-pause" flagUsage:"Click to pause/resume playback"`
	NoVideo           *bool    `flag:"no-video" flagUsage:"Start additional instances without video"`
	ReSeekSrc         *bool    `flag:"re-seek-src" flagUsage:"Re-seek source player for precise sync"`
	Debug             bool     `flag:"debug" flagUsage:"Debug mode"`
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

	flagSet := flago.NewFlagSet("", flag.ExitOnError)

	var ignoredFields []any
	//goland:noinspection GoBoolExpressions
	if !static_features.ClickPause {
		ignoredFields = append(ignoredFields, &args.ClickPause)
	}

	if err = flagSet.StructVar(&args, ignoredFields...); err != nil {
		return args, err
	}
	err = flagSet.Parse(os.Args[1:])
	return args, err
}
