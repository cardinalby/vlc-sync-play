package vlc_path

import "os/exec"

func getDefaultVlcBinPath() (string, error) {
	vlcPath, err := exec.LookPath("vlc")
	if err != nil {
		return "", err
	}
	return vlcPath, nil
}
