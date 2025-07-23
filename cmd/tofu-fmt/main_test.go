package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"pre-commit-hooks/internal/testutil"
	"testing"
)

// TestCheckOpenTofuInstalled tests the shared CheckOpenTofuInstalled function
func TestCheckOpenTofuInstalled(t *testing.T) {
	if _, err := exec.LookPath("tofu"); err == nil {
		if !testutil.CheckOpenTofuInstalled() {
			t.Error("CheckOpenTofuInstalled should return true when tofu is installed")
		}
	} else {
		t.Skip("Skipping positive test as tofu is not installed")
	}
	t.Log("Note: Unable to directly test the case where tofu is not installed")
}

func TestTofuFmt(t *testing.T) {
	// Skip test if tofu is not installed
	if _, err := exec.LookPath("tofu"); err != nil {
		t.Skip("Skipping test as tofu is not installed")
	}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "tofu_fmt_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create formatted and unformatted files
	formattedContent := `variable "example" {}`
	unformattedContent := `variable   "example"   {}` // Extra spaces

	formattedFilePath := filepath.Join(tempDir, "formatted.tf")
	unformattedFilePath := filepath.Join(tempDir, "unformatted.tf")

	if err := os.WriteFile(formattedFilePath, []byte(formattedContent), 0644); err != nil {
		t.Fatalf("Failed to write formatted file: %v", err)
	}
	if err := os.WriteFile(unformattedFilePath, []byte(unformattedContent), 0644); err != nil {
		t.Fatalf("Failed to write unformatted file: %v", err)
	}

	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory to %s: %v", tempDir, err)
	}

	// Run tofu fmt
	fmtCmd := exec.Command("tofu", "fmt")
	if output, err := fmtCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to run tofu fmt: %v, output: %s", err, output)
	}

	// Verify unformatted file is now formatted
	result, err := os.ReadFile(unformattedFilePath)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", unformattedFilePath, err)
	}
	if string(result) != formattedContent {
		t.Errorf("File was not formatted correctly. Got: %q, Want: %q", string(result), formattedContent)
	}
}
