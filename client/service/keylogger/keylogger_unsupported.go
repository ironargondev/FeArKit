//go:build !windows && !linux

package keylogger

import "errors"

func StartKeylogger() error {
	return errors.New(`${i18n|COMMON.OPERATION_NOT_SUPPORTED}`)
}

func GetKeyboardLayout() (string) {
	return "0"
}