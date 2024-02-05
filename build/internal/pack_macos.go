package commands

import (
	"os"
	"path/filepath"

	"github.com/cardinalby/vlc-sync-play/build/internal/consts"
	"github.com/cardinalby/vlc-sync-play/build/internal/docker"
	"github.com/cardinalby/vlc-sync-play/build/internal/fsutil"
	"github.com/cardinalby/vlc-sync-play/build/internal/macos"
	"golang.org/x/sync/errgroup"
)

func PackMacosBundles(
	bundles map[consts.Arch]macos.BundlePath,
	srcPsdIconPath string,
	plistSrcPath string,
	bundleIconSetFileName string,
	tmpDir string,
) error {
	tmpPngIconFilePath := filepath.Join(tmpDir, "icon.png")
	if err := docker.ImageMagicPsdToPng(srcPsdIconPath, tmpPngIconFilePath); err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(tmpPngIconFilePath)
	}()

	plistRelPath := filepath.Join("Info.plist")

	errGroup := errgroup.Group{}
	for _, bundle := range bundles {
		bundle := bundle
		errGroup.Go(func() error {
			if err := generateIconSet(
				tmpPngIconFilePath, filepath.Join(bundle.GetResourcesPath(), bundleIconSetFileName),
			); err != nil {
				return err
			}
			return fsutil.CopyFile(plistSrcPath, filepath.Join(bundle.GetContentsPath(), plistRelPath))
		})
	}
	return errGroup.Wait()
}
