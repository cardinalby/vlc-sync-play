package vlc_path

import (
	"os"
)

func GetDefaultVlcBinPath() (string, error) {
	p, err := getDefaultVlcBinPath()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(p); err != nil {
		return "", err
	}
	return p, nil
}
