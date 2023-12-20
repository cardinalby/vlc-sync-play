package instance

import (
	"errors"
	"os"
	"runtime"
)

func GetDefaultVlcBinPath() (string, error) {
	var possiblePaths []string

	switch runtime.GOOS {
	case "darwin":
		possiblePaths = []string{"/Applications/VLC.app/Contents/MacOS/VLC"}
	case "windows":
		possiblePaths = []string{
			os.Getenv("ProgramFiles") + "\\VideoLAN\\VLC\\vlc.exe",
			os.Getenv("ProgramFiles(x86)") + "\\VideoLAN\\VLC\\vlc.exe",
		}
	default:
		possiblePaths = []string{"vlc"}
	}

	for _, p := range possiblePaths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", errors.New("VLC not found")
}
