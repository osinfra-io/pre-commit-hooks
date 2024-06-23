package main

import (
	"bytes"
	"log"
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

	if err := os.WriteFile(formattedFilePath, []byte(formattedContent), 0644); err != nil {
		log.Fatalf("Failed to write formatted content to file: %v", err)
	}

	if err := os.WriteFile(unformattedFilePath, []byte(unformattedContent), 0644); err != nil {
		log.Fatalf("Failed to write unformatted content to file: %v", err)
	}

	// Change working directory to tempDir
	originalWd, _ := os.Getwd()

	if err := os.Chdir(tempDir); err != nil {
		log.Fatalf("Failed to change directory to %s: %v", tempDir, err)
	}

	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			log.Fatalf("Failed to restore original working directory to %s: %v", originalWd, err)
		}
	}()

	// Step 3: Capture output
	originalStdout := os.Stdout // Keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()

	w.Close()
	os.Stdout = originalStdout // Restore original stdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		log.Fatalf("Failed to read from reader: %v", err)
	}
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
