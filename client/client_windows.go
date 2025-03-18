//go:build windows

package main

import (
	"syscall"
)

func HideConsoleWindow() {
	// Hide the console window on Windows.
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	user32 := syscall.NewLazyDLL("user32.dll")
	getConsoleWindow := kernel32.NewProc("GetConsoleWindow")
	showWindow := user32.NewProc("ShowWindow")

	hwnd, _, _ := getConsoleWindow.Call()
	const SW_HIDE = 0
	showWindow.Call(hwnd, uintptr(SW_HIDE))
}