//go:build !windows && !linux

package executable

import "errors"

func DownloadAndExecute(url string, path string) error {
	return errors.New(`${i18n|COMMON.OPERATION_NOT_SUPPORTED}`)
}

func LoadElf(elf []byte, binaryPath string) error {
	return errors.New(`${i18n|COMMON.OPERATION_NOT_SUPPORTED}`)
}
