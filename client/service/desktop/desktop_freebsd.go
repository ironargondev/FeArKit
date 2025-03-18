//go:build freebsd

package desktop

import (
	"errors"
	"FeArKit/modules"
)

func Capture(bridge string) error {
	return errors.New(`${i18n|COMMON.OPERATION_NOT_SUPPORTED}`)
}

func Init(bridge string) error {
	return errors.New(`${i18n|COMMON.OPERATION_NOT_SUPPORTED}`)
}

func InitDesktop(pack modules.Packet) error {
	return errors.New(`${i18n|COMMON.OPERATION_NOT_SUPPORTED}`)
}

func PingDesktop(pack modules.Packet) error {
	return errors.New(`${i18n|COMMON.OPERATION_NOT_SUPPORTED}`)
}

func KillDesktop(pack modules.Packet) error {
	return errors.New(`${i18n|COMMON.OPERATION_NOT_SUPPORTED}`)
}

func GetDesktop(pack modules.Packet) error {
	return errors.New(`${i18n|COMMON.OPERATION_NOT_SUPPORTED}`)
}