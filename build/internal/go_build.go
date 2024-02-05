package commands

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/cardinalby/vlc-sync-play/build/internal/consts"
	"github.com/cardinalby/vlc-sync-play/build/internal/fsutil"
	xgolib "github.com/cardinalby/xgo-as-library"
	"golang.org/x/sync/errgroup"
)

type Target struct {
	Os   consts.Os
	Arch consts.Arch
}

func (t Target) String() string {
	return fmt.Sprintf("%s/%s", t.Os, t.Arch)
}

type Targets []Target

func (t Targets) Strings() []string {
	res := make([]string, len(t))
	for i, target := range t {
		res[i] = target.String()
	}
	return res
}

// ldFlags -> targets
type buildOptions map[string]Targets

func getLdFlags(target Target) string {
	if target.Os == consts.OsWindows {
		return "-H=windowsgui"
	}
	return ""
}

func getXgoOutBinPath(
	binTmpDir string,
	outPrefix string,
	target Target,
) string {
	res := path.Join(binTmpDir, fmt.Sprintf("%s-%s-%s", outPrefix, target.Os, target.Arch))
	if target.Os == consts.OsWindows {
		res += ".exe"
	}
	return res
}

func GoBuild(
	rootPath string,
	cmdPkgPath string,
	outPaths map[Target]string, // "os/arch" -> out path
	tmpDir string,
	logger *log.Logger,
) error {
	options := make(buildOptions)
	for target := range outPaths {
		ldFlags := getLdFlags(target)
		options[ldFlags] = append(options[ldFlags], target)
	}

	binTmpDir := path.Join(tmpDir, "bin")
	defer func() {
		if err := os.RemoveAll(binTmpDir); err != nil {
			logger.Printf("error removing tmp '%s': %v", binTmpDir, err)
		}
	}()

	errGr := errgroup.Group{}
	outPrefix := "vlcsp"

	for ldFlags, targets := range options {
		targets := targets
		args := xgolib.Args{
			Repository: rootPath,
			SrcPackage: cmdPkgPath,
			OutFolder:  binTmpDir,
			OutPrefix:  outPrefix,
			Build: xgolib.BuildArgs{
				LdFlags: ldFlags,
			},
			Targets: targets.Strings(),
		}
		errGr.Go(func() error {
			if err := xgolib.StartBuild(args, logger); err != nil {
				return err
			}
			for _, target := range targets {
				outBinPath := getXgoOutBinPath(binTmpDir, outPrefix, target)
				if err := fsutil.RenameFile(outBinPath, outPaths[target]); err != nil {
					return err
				}
			}
			return nil
		})
	}
	return errGr.Wait()
}
