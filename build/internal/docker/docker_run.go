package docker

import (
	"errors"
	"fmt"
	"os/exec"
)

func RunImage(image string, volumes map[string]string, args []string) error {
	allArgs := append([]string{"run", "--rm"})
	var volumeArgs []string
	for src, dst := range volumes {
		volumeArgs = append(volumeArgs, "-v", fmt.Sprintf(`%s:%s`, src, dst))
	}
	allArgs = append(allArgs, volumeArgs...)
	allArgs = append(allArgs, image)
	allArgs = append(allArgs, args...)

	cmd := exec.Command(
		"docker",
		allArgs...,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Join(
			fmt.Errorf("%w. Command: docker %v", err, allArgs),
			errors.New(string(out)))
	}
	return nil
}
