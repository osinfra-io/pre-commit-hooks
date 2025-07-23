package tofu_validate

import (
	"os"
	"os/exec"
	"testing"
)

func TestCheckOpenTofuInstalled(t *testing.T) {
	_ = CheckOpenTofuInstalled()
}

func TestRunTofuValidate_ValidConfig(t *testing.T) {
	if !CheckOpenTofuInstalled() {
		t.Skip("Skipping: tofu is not installed")
	}
	tempDir, err := os.MkdirTemp("", "validate_test_valid")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create valid .tf file
	filePath := tempDir + "/main.tf"
	valid := `terraform {
  required_version = ">= 1.0.0"
}

resource "null_resource" "example" {}`
	if err := os.WriteFile(filePath, []byte(valid), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Run tofu init first
	initCmd := []string{"init", "-input=false", "--backend=false"}
	cmd := exec.Command("tofu", initCmd...)
	cmd.Dir = tempDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Expected tofu init to succeed for valid config, but got error: %v, output: %s", err, out)
	}

	output, err := RunTofuValidate(tempDir, nil)
	if err != nil {
		t.Fatalf("Expected valid config to pass validation, but got error: %v, output: %s", err, output)
	}
}

func TestRunTofuValidate_InvalidConfig(t *testing.T) {
	if !CheckOpenTofuInstalled() {
		t.Skip("Skipping: tofu is not installed")
	}
	tempDir, err := os.MkdirTemp("", "validate_test_invalid")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create invalid .tf file
	filePath := tempDir + "/main.tf"
	invalid := `terraform {
  required_version = ">= 1.0.0"
  # Missing closing brace intentionally`
	if err := os.WriteFile(filePath, []byte(invalid), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Run tofu init first
	initCmd := []string{"init", "-input=false", "--backend=false"}
	cmd := exec.Command("tofu", initCmd...)
	cmd.Dir = tempDir
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("Expected tofu init to fail due to invalid config, but it succeeded. Output: %s", out)
	}
	// If tofu init fails as expected, test passes

	output, err := RunTofuValidate(tempDir, nil)
	if err == nil {
		t.Fatalf("Expected invalid config to fail validation, but got no error. Output: %s", output)
	}
}
