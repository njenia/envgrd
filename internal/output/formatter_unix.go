//go:build !windows
// +build !windows

package output

// enableANSI returns true on Unix-like systems if stdout is a terminal
// Colors are supported by default on Unix terminals
func enableANSI() bool {
	return true
}

