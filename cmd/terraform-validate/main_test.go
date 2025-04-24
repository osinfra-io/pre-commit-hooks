package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestCheckTerraformInstalled tests the checkTerraformInstalled function
func TestCheckTerraformInstalled(t *testing.T) {
	// Test case 1: Terraform exists
	// This test relies on the real function and will be skipped if terraform is not available
	if _, err := exec.LookPath("terraform"); err == nil {
		if !checkTerraformInstalled() {
			t.Error("checkTerraformInstalled should return true when terraform is installed")
		}
	} else {
		t.Skip("Skipping positive test as terraform is not installed")
	}

	// Test case 2: We can't easily test the negative case without modifying the PATH
	// or using a custom function, so we'll just document this limitation
	t.Log("Note: Unable to directly test the case where terraform is not installed")
}

// TestValidTerraformConfig tests that a valid config passes validation
func TestValidTerraformConfig(t *testing.T) {
	// Skip this test if terraform is not installed
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("Skipping test as terraform is not installed")
	}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "terraform_validate_test")
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

	// Run terraform init
	initCmd := exec.Command("terraform", "init")
	if initOutput, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to initialize terraform: %v, output: %s", err, initOutput)
	}

	// Run terraform validate directly to confirm the configuration is valid
	validateCmd := exec.Command("terraform", "validate")
	if validateOutput, err := validateCmd.CombinedOutput(); err != nil {
		t.Fatalf("Expected valid config to pass validation, but it failed: %v, output: %s", err, validateOutput)
	}
}

// TestInvalidTerraformConfig tests that an invalid config fails validation
func TestInvalidTerraformConfig(t *testing.T) {
	// Skip this test if terraform is not installed
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("Skipping test as terraform is not installed")
	}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "terraform_validate_test")
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
	invalidContent := `terraform {
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
	// Run terraform validate directly to confirm the configuration is invalid
	validateCmd := exec.Command("terraform", "validate")
	output, err := validateCmd.CombinedOutput()
	if err == nil {
		t.Fatalf("Expected invalid config to fail validation, but it passed. Output: %s", output)
	}
}
