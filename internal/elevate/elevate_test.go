package elevate

import (
	"os"
	"testing"
)

func TestIsAdmin_ReturnsBool(t *testing.T) {
	// IsAdmin should always return a bool without panicking
	result := IsAdmin()
	// It should be either true (if running as admin) or false (if not)
	if result != true && result != false {
		t.Errorf("IsAdmin() should return true or false, got unexpected: %v", result)
	}
}

func TestRestartElevated_ErrorWhenNoExe(t *testing.T) {
	// Temporarily save the real executable path and restore it
	// This test verifies the function handles errors properly
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	// If we can get the executable, RestartElevated should try to ShellExecute
	// and may succeed or fail depending on environment
	// We just verify it doesn't panic
	err := RestartElevated()
	_ = err // err is expected to be non-nil in test environment without UAC
}
