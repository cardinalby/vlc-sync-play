package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
)

func RenameFile(src, dst string) error {
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("error creating dst dir: %w", err)
	}
	return os.Rename(src, dst)
}
