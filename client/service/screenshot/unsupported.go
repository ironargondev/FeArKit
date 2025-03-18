//go:build !linux && !windows && !darwin && freebsd

package screenshot

import "errors"

func GetScreenshot(bridge string) error {
	return errors.New(`${i18n|COMMON.OPERATION_NOT_SUPPORTED}`)
}
