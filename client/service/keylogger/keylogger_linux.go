//go:build linux

package keylogger

import (
	"os/exec"
	"strings"
	"errors"
)
func StartKeylogger() error {
	return errors.New(`${i18n|COMMON.OPERATION_NOT_SUPPORTED}`)
}

func GetKeyboardLayout() (string) {
	cmdOutput, err := exec.Command("setxkbmap", "-query").Output()
	if err != nil {
		return "us" // fallback layout
	}
	for _, line := range strings.Split(string(cmdOutput), "\n") {
		if strings.HasPrefix(strings.ToLower(line), "layout:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return "us"
}