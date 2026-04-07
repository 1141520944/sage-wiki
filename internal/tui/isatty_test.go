package tui

import (
	"os"
	"testing"
)

func TestIsInteractive(t *testing.T) {
	// Verify IsInteractive matches actual stdin state
	fi, err := os.Stdin.Stat()
	if err != nil {
		t.Fatalf("failed to stat stdin: %v", err)
	}
	expected := fi.Mode()&os.ModeCharDevice != 0
	got := IsInteractive()
	if got != expected {
		t.Errorf("IsInteractive() = %v, expected %v", got, expected)
	}
}
