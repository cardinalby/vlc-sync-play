package commands

import (
	"fmt"
	"image"
	"os"
	"path/filepath"

	"github.com/jackmordaunt/icns"
)

func generateIconSet(srcIconPath, dstIconSetPath string) error {
	pngf, err := os.Open(srcIconPath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer func() {
		_ = pngf.Close()
	}()
	srcImg, _, err := image.Decode(pngf)
	if err != nil {
		return fmt.Errorf("error decoding png: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(dstIconSetPath), 0755); err != nil {
		return fmt.Errorf("error creating output dir: %w", err)
	}
	dest, err := os.Create(dstIconSetPath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}

	if err := icns.Encode(dest, srcImg); err != nil {
		_ = dest.Close()
		return fmt.Errorf("error encoding icns: %w", err)
	}
	return dest.Close()
}
