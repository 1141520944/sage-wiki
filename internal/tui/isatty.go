package tui

import "os"

// IsInteractive returns true when stdin is a terminal (not piped).
func IsInteractive() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
