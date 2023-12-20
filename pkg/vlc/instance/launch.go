package instance

import (
	"os"
	"os/exec"

	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic/protocols"
)

type Launcher func(filePaths []string, noVideo bool) (Instance, error)

func GetLauncher(vlcPath string, apiProtocol protocols.ApiProtocol) Launcher {
	return func(filePaths []string, noVideo bool) (Instance, error) {
		apiClient, err := protocols.NewLocalBasicApiClient(apiProtocol)
		if err != nil {
			return Instance{}, err
		}

		args := apiClient.GetLaunchArgs()
		if noVideo {
			args = append(args, "--no-video")
		}
		args = append(args, filePaths...)
		cmd := exec.Command(vlcPath, args...)
		if workingDir, err := os.Getwd(); err == nil {
			cmd.Dir = workingDir
		}
		cmd.Env = os.Environ()
		return Instance{
			ApiClient: apiClient,
			Cmd:       cmd,
		}, cmd.Start()
	}
}
