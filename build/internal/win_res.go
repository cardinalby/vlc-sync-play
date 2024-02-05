package commands

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/akavel/rsrc/rsrc"
	"github.com/cardinalby/vlc-sync-play/build/internal/docker"
)

func GenerateWindowsSysoFile(
	srcPsdIconPath string,
	manifestPath string,
	tmpDir string,
	outputPath string,
	logger *log.Logger,
) error {
	tmpIconPath := filepath.Join(tmpDir, "tmp-vlc-sync-play-icon.ico")

	err := docker.ImageMagicPsdToIco(srcPsdIconPath, tmpIconPath, []int{16, 32, 48, 64, 128, 256})
	if err != nil {
		return fmt.Errorf("error generating temporary ico file: %w", err)
	}
	defer func() {
		if err := os.Remove(tmpIconPath); err != nil {
			logger.Printf("error removing temporary '%s': %v", tmpIconPath, err)
		}
	}()

	return rsrc.Embed(outputPath, "amd64", manifestPath, tmpIconPath)
}
