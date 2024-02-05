package macos

import "path/filepath"

type BundlePath string

func (p BundlePath) GetContentsPath() string {
	return filepath.Join(string(p), "Contents")
}

func (p BundlePath) GetBinDir() string {
	return filepath.Join(p.GetContentsPath(), "MacOS")
}

func (p BundlePath) GetPlistPath() string {
	return filepath.Join(p.GetContentsPath(), "Info.plist")
}

func (p BundlePath) GetResourcesPath() string {
	return filepath.Join(p.GetContentsPath(), "Resources")
}
