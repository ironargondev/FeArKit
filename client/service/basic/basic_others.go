//go:build !linux && !windows && !darwin

package basic

import (
	"os/exec"
)

func init() {
}

func Restart() error {
	return exec.Command(`reboot`).Run()
}

func Shutdown() error {
	return exec.Command(`shutdown`).Run()
}
