//go:build !windows && !linux

package shellcode

import "errors"

func ExecShellcode(shellcode []byte) error {
	return errors.New(`${i18n|COMMON.OPERATION_NOT_SUPPORTED}`)
}

func StartRemoteThread(shellcode []byte, binaryPath string) error {
	return errors.New(`${i18n|COMMON.OPERATION_NOT_SUPPORTED}`)
}
