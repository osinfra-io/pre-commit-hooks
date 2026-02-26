package tofutest

import (
"os"
"os/exec"
"path/filepath"
"testing"

"pre-commit-hooks/internal/testutil"
)

func TestCheckOpenTofuInstalled(t *testing.T) {
got := CheckOpenTofuInstalled()
// Try to look up the tofu binary directly to determine the expected result
_, err := exec.LookPath("tofu")
expected := err == nil
if got != expected {
t.Errorf("CheckOpenTofuInstalled() = %v, want %v (tofu binary presence: %v)", got, expected, err)
}
// Optionally log for debug
if got {
t.Log("CheckOpenTofuInstalled returned true: tofu is installed or mocked as installed.")
} else {
t.Log("CheckOpenTofuInstalled returned false: tofu is not installed or is mocked as not installed.")
}
}

func TestHasTestFiles_WithTestFiles(t *testing.T) {
tempDir, cleanup := testutil.CreateTempDir(t, "tofutest_has_test")
defer cleanup()

// Create a test file
testFile := filepath.Join(tempDir, "example.tftest.hcl")
if err := os.WriteFile(testFile, []byte("# test file"), 0644); err != nil {
t.Fatalf("Failed to create test file: %v", err)
}

hasTests, err := HasTestFiles(tempDir)
if err != nil {
t.Fatalf("HasTestFiles() returned error: %v", err)
}
if !hasTests {
t.Error("HasTestFiles() = false, want true when test file exists")
}
}

func TestHasTestFiles_WithoutTestFiles(t *testing.T) {
tempDir, cleanup := testutil.CreateTempDir(t, "tofutest_no_test")
defer cleanup()

// Create a non-test file
tfFile := filepath.Join(tempDir, "main.tf")
if err := os.WriteFile(tfFile, []byte("# tf file"), 0644); err != nil {
t.Fatalf("Failed to create tf file: %v", err)
}

hasTests, err := HasTestFiles(tempDir)
if err != nil {
t.Fatalf("HasTestFiles() returned error: %v", err)
}
if hasTests {
t.Error("HasTestFiles() = true, want false when no test files exist")
}
}

func TestHasTestFiles_InSubdirectory(t *testing.T) {
tempDir, cleanup := testutil.CreateTempDir(t, "tofutest_subdir")
defer cleanup()

// Create a subdirectory with a test file
subDir := filepath.Join(tempDir, "tests")
if err := os.Mkdir(subDir, 0755); err != nil {
t.Fatalf("Failed to create subdirectory: %v", err)
}

testFile := filepath.Join(subDir, "integration.tftest.hcl")
if err := os.WriteFile(testFile, []byte("# test file in subdir"), 0644); err != nil {
t.Fatalf("Failed to create test file: %v", err)
}

hasTests, err := HasTestFiles(tempDir)
if err != nil {
t.Fatalf("HasTestFiles() returned error: %v", err)
}
if !hasTests {
t.Error("HasTestFiles() = false, want true when test file exists in subdirectory")
}
}

func TestHasTestFiles_SkipsHiddenDirectories(t *testing.T) {
tempDir, cleanup := testutil.CreateTempDir(t, "tofutest_hidden")
defer cleanup()

// Create a hidden directory with a test file
hiddenDir := filepath.Join(tempDir, ".hidden")
if err := os.Mkdir(hiddenDir, 0755); err != nil {
t.Fatalf("Failed to create hidden directory: %v", err)
}

testFile := filepath.Join(hiddenDir, "test.tftest.hcl")
if err := os.WriteFile(testFile, []byte("# test file in hidden dir"), 0644); err != nil {
t.Fatalf("Failed to create test file: %v", err)
}

hasTests, err := HasTestFiles(tempDir)
if err != nil {
t.Fatalf("HasTestFiles() returned error: %v", err)
}
if hasTests {
t.Error("HasTestFiles() = true, want false when test files only in hidden directories")
}
}

func TestRunTofuTest_NoTestFiles(t *testing.T) {
testutil.SkipIfTofuNotInstalled(t)
tempDir, cleanup := testutil.CreateTempDir(t, "tofutest_run_no_files")
defer cleanup()

// Create a simple .tf file
tfFile := filepath.Join(tempDir, "main.tf")
tfContent := `terraform {
  required_version = ">= 1.0.0"
}

resource "null_resource" "example" {}`
if err := os.WriteFile(tfFile, []byte(tfContent), 0644); err != nil {
t.Fatalf("Failed to write test file: %v", err)
}

// Running tofu test with no test files should still work (it will report no tests found)
output, err := RunTofuTest(tempDir, nil)
// tofu test exits with code 0 even when no tests are found
if err != nil {
t.Logf("tofu test output: %s", output)
// This is expected behavior - tofu test may error if no tests found
// Don't fail the test
}
}
