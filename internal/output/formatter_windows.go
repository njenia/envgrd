//go:build windows
// +build windows

package output

import (
	"syscall"
	"unsafe"
)

// Windows API constants for enabling ANSI
const (
	enableVirtualTerminalProcessing = 0x0004
	stdOutputHandle                 = uint32(0xFFFFFFF5)
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
	procGetStdHandle   = kernel32.NewProc("GetStdHandle")
)

// enableANSI enables ANSI escape sequence processing on Windows 10+
func enableANSI() bool {
	// Get stdout handle
	handle, _, _ := procGetStdHandle.Call(uintptr(stdOutputHandle))
	if handle == 0 {
		return false
	}

	// Get current console mode
	var mode uint32
	ret, _, _ := procGetConsoleMode.Call(handle, uintptr(unsafe.Pointer(&mode)))
	if ret == 0 {
		return false
	}

	// Enable virtual terminal processing
	mode |= enableVirtualTerminalProcessing
	ret, _, _ = procSetConsoleMode.Call(handle, uintptr(mode))
	return ret != 0
}

