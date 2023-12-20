package args

import (
	"errors"
	"flag"
	"os"
	"time"

	"github.com/cardinalby/vlc-sync-play/pkg/vlc/instance"
)

type CmdLineArgs struct {
	VlcPath           string
	FilePaths         []string
	InstancesNumber   int
	PollingIntervalMs int
}

func (args CmdLineArgs) GetPollingInterval() time.Duration {
	return time.Duration(args.PollingIntervalMs) * time.Millisecond
}

func ParseCmdLineArgs() (a CmdLineArgs, err error) {
	flagSet := flag.NewFlagSet("", flag.ExitOnError)
	flagSet.IntVar(&a.InstancesNumber, "n", 2, "Number of VLC instances")
	flagSet.IntVar(&a.PollingIntervalMs, "i", 100, "Polling interval ms")
	flagSet.StringVar(&a.VlcPath, "vlc", "", "Path to VLC executable")
	if err = flagSet.Parse(os.Args[1:]); err != nil {
		return a, err
	}
	a.FilePaths = flagSet.Args()
	// TODO
	a.FilePaths = []string{"/Users/cardinalby/Movies/Ofis.1.sezon.6.serija.2005.x264.BDRip.720p-kernlas.mkv"}

	if a.VlcPath == "" {
		if defaultPath, err := instance.GetDefaultVlcBinPath(); err == nil {
			a.VlcPath = defaultPath
		}
	}

	if a.InstancesNumber < 2 {
		return a, errors.New("'n' should be at least 2")
	}

	return a, nil
}
