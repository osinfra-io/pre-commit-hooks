package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenTofuFormat(t *testing.T) {
	// Step 1: Create temporary directory
	tempDir, err := os.MkdirTemp("", ".go_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Cleanup

	// Step 2: Create OpenTofu files
	formattedContent := `variable "example" {}`
	unformattedContent := `variable   "example"   {}` // Extra spaces

	formattedFilePath, err := filepath.Abs(filepath.Join(tempDir, "formatted.tf"))
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}
	unformattedFilePath, err := filepath.Abs(filepath.Join(tempDir, "unformatted.tf"))
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	if err := os.WriteFile(formattedFilePath, []byte(formattedContent), 0644); err != nil {
		t.Fatalf("Failed to write formatted file: %v", err)
	}
	if err := os.WriteFile(unformattedFilePath, []byte(unformattedContent), 0644); err != nil {
		t.Fatalf("Failed to write unformatted file: %v", err)
	}

	// Ensure tempDir is an absolute path
	absTempDir, err := filepath.Abs(tempDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path for %s: %v", tempDir, err)
	}

	// Get the current working directory as an absolute path
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	// Change working directory to absTempDir
	if err := os.Chdir(absTempDir); err != nil {
		t.Fatalf("Failed to change directory to %s: %v", absTempDir, err)
	}

	// Defer the restoration of the original working directory
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Fatalf("Failed to restore original working directory to %s: %v", originalWd, err)
		}
	}()

	// Step 3: Capture output
	originalStdout := os.Stdout // Keep backup of the real stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	main()

	w.Close()
	os.Stdout = originalStdout // Restore original stdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from reader: %v", err)
	}
	output := buf.String()

	// Step 4: Verify results
	if !strings.Contains(output, "Files formatted successfully with tofu fmt.") {
		t.Errorf("Expected success message not found in output")
	}

	// Verify unformatted file is now formatted
	unformattedResult, err := os.ReadFile(unformattedFilePath)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", unformattedFilePath, err)
	}
	if string(unformattedResult) != formattedContent {
		t.Errorf("File was not formatted correctly")
	}
}
