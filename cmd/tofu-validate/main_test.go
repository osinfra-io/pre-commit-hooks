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

// TestValidOpenTofuConfig tests that a valid config passes validation
func TestValidOpenTofuConfig(t *testing.T) {
	// Skip this test if tofu is not installed
	if _, err := exec.LookPath("tofu"); err != nil {
		t.Skip("Skipping test as tofu is not installed")
	}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "tofu_validate_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create valid configuration
	validDir := filepath.Join(tempDir, "valid")
	if err := os.Mkdir(validDir, 0755); err != nil {
		t.Fatalf("Failed to create valid config directory: %v", err)
	}

	// Write valid configuration
	validContent := `terraform {
  required_version = ">= 1.0.0"
}

resource "local_file" "example" {
  content  = "example content"
  filename = "${path.module}/example.txt"
}`

	validFilePath := filepath.Join(validDir, "main.tf")
	if err := os.WriteFile(validFilePath, []byte(validContent), 0644); err != nil {
		t.Fatalf("Failed to write valid file: %v", err)
	}

	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Change to test directory
	if err := os.Chdir(validDir); err != nil {
		t.Fatalf("Failed to change directory to %s: %v", validDir, err)
	}

	// Run tofu init
	initCmd := exec.Command("tofu", "init")
	if initOutput, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to initialize tofu: %v, output: %s", err, initOutput)
	}

	// Run tofu validate directly to confirm the configuration is valid
	validateCmd := exec.Command("tofu", "validate")
	if validateOutput, err := validateCmd.CombinedOutput(); err != nil {
		t.Fatalf("Expected valid config to pass validation, but it failed: %v, output: %s", err, validateOutput)
	}
}

// TestInvalidOpenTofuConfig tests that an invalid config fails validation
func TestInvalidOpenTofuConfig(t *testing.T) {
	// Skip this test if tofu is not installed
	if _, err := exec.LookPath("tofu"); err != nil {
		t.Skip("Skipping test as tofu is not installed")
	}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "tofu_validate_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create invalid configuration
	invalidDir := filepath.Join(tempDir, "invalid")
	if err := os.Mkdir(invalidDir, 0755); err != nil {
		t.Fatalf("Failed to create invalid config directory: %v", err)
	}

	// Write invalid configuration (syntax error)
	invalidContent := `tofu {
  required_version = ">= 1.0.0"
  # Missing closing brace intentionally`

	invalidFilePath := filepath.Join(invalidDir, "main.tf")
	if err := os.WriteFile(invalidFilePath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to write invalid file: %v", err)
	}

	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Change to test directory
	if err := os.Chdir(invalidDir); err != nil {
		t.Fatalf("Failed to change directory to %s: %v", invalidDir, err)
	}
	// Run tofu validate directly to confirm the configuration is invalid
	validateCmd := exec.Command("tofu", "validate")
	output, err := validateCmd.CombinedOutput()
	if err == nil {
		t.Fatalf("Expected invalid config to fail validation, but it passed. Output: %s", output)
	}
}
