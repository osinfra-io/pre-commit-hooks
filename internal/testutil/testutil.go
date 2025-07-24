package testutil

import (
	"os"
	"os/exec"
	"testing"
)

// CheckOpenTofuInstalled returns true if the 'tofu' binary is found in PATH.
func CheckOpenTofuInstalled() bool {
	_, err := exec.LookPath("tofu")
	return err == nil
}

// SkipIfTofuNotInstalled skips the test if tofu is not installed
func SkipIfTofuNotInstalled(t *testing.T) {
	if !CheckOpenTofuInstalled() {
		t.Skip("Skipping test as tofu is not installed")
	}
}

// CreateTempDir creates a temp directory and returns its path and a cleanup function
func CreateTempDir(t *testing.T, prefix string) (string, func()) {
	dir, err := os.MkdirTemp("", prefix)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	cleanup := func() { os.RemoveAll(dir) }
	return dir, cleanup
}

// RestoreWorkingDir changes to a new directory and returns a function to restore the original
func RestoreWorkingDir(t *testing.T, newDir string) func() {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	if err := os.Chdir(newDir); err != nil {
		t.Fatalf("Failed to change directory to %s: %v", newDir, err)
	}
	return func() {
		os.Chdir(origDir)
	}
}

// RunCmd runs a command and returns its combined output and error
func RunCmd(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
