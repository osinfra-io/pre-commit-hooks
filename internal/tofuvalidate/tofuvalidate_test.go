package tofu_validate

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"pre-commit-hooks/internal/testutil"
)

func TestCheckOpenTofuInstalled(t *testing.T) {
	_ = CheckOpenTofuInstalled()
	got := CheckOpenTofuInstalled()
	if got {
		t.Log("CheckOpenTofuInstalled returned true: tofu is installed or mocked as installed.")
	} else {
		t.Log("CheckOpenTofuInstalled returned false: tofu is not installed or is mocked as not installed.")
	}
}

func TestRunTofuValidate_ValidConfig(t *testing.T) {
	testutil.SkipIfTofuNotInstalled(t)
	tempDir, cleanup := testutil.CreateTempDir(t, "validate_test_valid")
	defer cleanup()

	// Create valid .tf file
	filePath := filepath.Join(tempDir, "main.tf")
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
	testutil.SkipIfTofuNotInstalled(t)
	tempDir, cleanup := testutil.CreateTempDir(t, "validate_test_invalid")
	defer cleanup()

	filePath := filepath.Join(tempDir, "main.tf")
	invalid := `terraform {
  required_version = ">= 1.0.0"
  # Missing closing brace intentionally`
	if err := os.WriteFile(filePath, []byte(invalid), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Run tofu init first (expected to fail, but test continues to run validation and checks for an error there; test does not pass solely on init failure)
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
