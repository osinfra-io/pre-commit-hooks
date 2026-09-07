package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"pre-commit-hooks/internal/testutil"
)

func TestRunTofuTestCLI_TofuNotInstalled(t *testing.T) {
	checkInstalled := func() bool { return false }
	getwd := func() (string, error) { return "/fake", nil }
	hasTestFiles := func(string) (bool, error) { return true, nil }
	runTest := func(string, []string) (string, error) { return "", nil }
	printStatus := func(string, string) {}

	err := RunTofuTestCLI(nil, checkInstalled, getwd, hasTestFiles, runTest, printStatus)

	if err == nil {
		t.Error("Expected error when tofu not installed, got nil")
	}
}

func TestRunTofuTestCLI_GetwdError(t *testing.T) {
	checkInstalled := func() bool { return true }
	getwd := func() (string, error) { return "", errors.New("getwd failed") }
	hasTestFiles := func(string) (bool, error) { return true, nil }
	runTest := func(string, []string) (string, error) { return "", nil }
	printStatus := func(string, string) {}

	err := RunTofuTestCLI(nil, checkInstalled, getwd, hasTestFiles, runTest, printStatus)

	if err == nil {
		t.Error("Expected error when getwd fails, got nil")
	}
}

func TestRunTofuTestCLI_NoTestFiles(t *testing.T) {
	checkInstalled := func() bool { return true }
	getwd := func() (string, error) { return "/fake", nil }
	hasTestFiles := func(string) (bool, error) { return false, nil }
	runTest := func(string, []string) (string, error) { return "", nil }
	printStatus := func(string, string) {}

	err := RunTofuTestCLI(nil, checkInstalled, getwd, hasTestFiles, runTest, printStatus)

	if err != nil {
		t.Errorf("Expected no error when no test files, got %v", err)
	}
}

func TestRunTofuTestCLI_HasTestFilesError(t *testing.T) {
	checkInstalled := func() bool { return true }
	getwd := func() (string, error) { return "/fake", nil }
	hasTestFiles := func(string) (bool, error) { return false, errors.New("walk error") }
	runTest := func(string, []string) (string, error) { return "", nil }
	printStatus := func(string, string) {}

	err := RunTofuTestCLI(nil, checkInstalled, getwd, hasTestFiles, runTest, printStatus)

	if err == nil {
		t.Error("Expected error when hasTestFiles fails, got nil")
	}
}

func TestRunTofuTestCLI_TestSuccess(t *testing.T) {
	checkInstalled := func() bool { return true }
	getwd := func() (string, error) { return "/fake", nil }
	hasTestFiles := func(string) (bool, error) { return true, nil }
	runTest := func(string, []string) (string, error) { return "All tests passed", nil }
	printStatus := func(string, string) {}

	err := RunTofuTestCLI(nil, checkInstalled, getwd, hasTestFiles, runTest, printStatus)

	if err != nil {
		t.Errorf("Expected no error when tests pass, got %v", err)
	}
}

func TestRunTofuTestCLI_TestFailure(t *testing.T) {
	rootCause := errors.New("test error")

	checkInstalled := func() bool { return true }
	getwd := func() (string, error) { return "/fake", nil }
	hasTestFiles := func(string) (bool, error) { return true, nil }
	runTest := func(string, []string) (string, error) { return "Test failed", rootCause }
	printStatus := func(string, string) {}

	err := RunTofuTestCLI(nil, checkInstalled, getwd, hasTestFiles, runTest, printStatus)

	if err == nil {
		t.Error("Expected error when tests fail, got nil")
	}
	if !errors.Is(err, rootCause) {
		t.Errorf("Expected returned error to wrap root cause, got: %v", err)
	}
}

func TestRunTofuTestCLI_ExtraArgs(t *testing.T) {
	var receivedArgs []string

	checkInstalled := func() bool { return true }
	getwd := func() (string, error) { return "/fake", nil }
	hasTestFiles := func(string) (bool, error) { return true, nil }
	runTest := func(dir string, args []string) (string, error) {
		receivedArgs = args
		return "Tests passed", nil
	}
	printStatus := func(string, string) {}

	extraArgs := []string{"-verbose", "-filter=TestFoo"}
	err := RunTofuTestCLI(extraArgs, checkInstalled, getwd, hasTestFiles, runTest, printStatus)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(receivedArgs) != 2 {
		t.Errorf("Expected 2 args, got %d", len(receivedArgs))
	}
	if len(receivedArgs) > 0 && receivedArgs[0] != "-verbose" {
		t.Errorf("Expected first arg to be -verbose, got %s", receivedArgs[0])
	}
	if len(receivedArgs) > 1 && receivedArgs[1] != "-filter=TestFoo" {
		t.Errorf("Expected second arg to be -filter=TestFoo, got %s", receivedArgs[1])
	}
}

func TestHasTestFiles_WithTestFiles(t *testing.T) {
	tempDir, cleanup := testutil.CreateTempDir(t, "tofutest-has-test")
	defer cleanup()

	testFile := filepath.Join(tempDir, "example.tftest.hcl")
	if err := os.WriteFile(testFile, []byte("# test file"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	got, err := hasTestFiles(tempDir)
	if err != nil {
		t.Fatalf("hasTestFiles() returned error: %v", err)
	}
	if !got {
		t.Error("hasTestFiles() = false, want true when test file exists")
	}
}

func TestHasTestFiles_WithoutTestFiles(t *testing.T) {
	tempDir, cleanup := testutil.CreateTempDir(t, "tofutest-no-test")
	defer cleanup()

	tfFile := filepath.Join(tempDir, "main.tf")
	if err := os.WriteFile(tfFile, []byte("# tf file"), 0o644); err != nil {
		t.Fatalf("Failed to create tf file: %v", err)
	}

	got, err := hasTestFiles(tempDir)
	if err != nil {
		t.Fatalf("hasTestFiles() returned error: %v", err)
	}
	if got {
		t.Error("hasTestFiles() = true, want false when no test files exist")
	}
}

func TestHasTestFiles_InSubdirectory(t *testing.T) {
	tempDir, cleanup := testutil.CreateTempDir(t, "tofutest-subdir")
	defer cleanup()

	subDir := filepath.Join(tempDir, "tests")
	if err := os.Mkdir(subDir, 0o755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	testFile := filepath.Join(subDir, "integration.tftest.hcl")
	if err := os.WriteFile(testFile, []byte("# test file in subdir"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	got, err := hasTestFiles(tempDir)
	if err != nil {
		t.Fatalf("hasTestFiles() returned error: %v", err)
	}
	if !got {
		t.Error("hasTestFiles() = false, want true when test file exists in subdirectory")
	}
}

func TestHasTestFiles_SkipsHiddenDirectories(t *testing.T) {
	tempDir, cleanup := testutil.CreateTempDir(t, "tofutest-hidden")
	defer cleanup()

	hiddenDir := filepath.Join(tempDir, ".hidden")
	if err := os.Mkdir(hiddenDir, 0o755); err != nil {
		t.Fatalf("Failed to create hidden directory: %v", err)
	}

	testFile := filepath.Join(hiddenDir, "test.tftest.hcl")
	if err := os.WriteFile(testFile, []byte("# test file in hidden dir"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	got, err := hasTestFiles(tempDir)
	if err != nil {
		t.Fatalf("hasTestFiles() returned error: %v", err)
	}
	if got {
		t.Error("hasTestFiles() = true, want false when test files only in hidden directories")
	}
}
