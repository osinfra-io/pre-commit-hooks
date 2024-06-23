package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTerraformFormat(t *testing.T) {
	// Step 1: Create temporary directory
	tempDir, err := os.MkdirTemp("", ".go_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Cleanup

	// Step 2: Create Terraform files
	formattedContent := `variable "example" {}`
	unformattedContent := `variable   "example"   {}` // Unformatted spaces

	formattedFilePath := filepath.Join(tempDir, "formatted.tf")
	unformattedFilePath := filepath.Join(tempDir, "unformatted.tf")

	os.WriteFile(formattedFilePath, []byte(formattedContent), 0644)
	os.WriteFile(unformattedFilePath, []byte(unformattedContent), 0644)

	// Change working directory to tempDir
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd) // Restore working directory

	// Step 3: Capture output
	originalStdout := os.Stdout // Keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()

	w.Close()
	os.Stdout = originalStdout // Restore original stdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Step 4: Verify results
	if !strings.Contains(output, "Files formatted successfully with terraform fmt.") {
		t.Errorf("Expected success message not found in output")
	}

	// Verify unformatted file is now formatted
	unformattedResult, _ := os.ReadFile(unformattedFilePath)
	if string(unformattedResult) != formattedContent {
		t.Errorf("File was not formatted correctly")
	}
}
