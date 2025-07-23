package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"pre-commit-hooks/internal/testutil"
	"strings"
	"testing"
)

func Test_runCmdInDir(t *testing.T) {
	out, err := runCmdInDir(".", []string{"nonexistentcmd"})
	if err == nil {
		t.Error("Expected error for nonexistent command")
	}
	if out == "" {
		t.Log("Output is empty as expected for nonexistent command")
	}
}

func Test_findDirsWithTfFiles_and_walkDirs(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "finddirs_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Mkdir(filepath.Join(tempDir, "sub1"), 0755)
	os.Mkdir(filepath.Join(tempDir, "sub2"), 0755)
	os.WriteFile(filepath.Join(tempDir, "sub1", "main.tf"), []byte("terraform {}"), 0644)
	os.WriteFile(filepath.Join(tempDir, "sub2", "other.txt"), []byte("not tf"), 0644)

	dirs := findDirsWithTfFiles(tempDir)
	found := false
	for _, d := range dirs {
		if strings.HasSuffix(d, "sub1") {
			found = true
		}
	}
	if !found {
		t.Error("Expected to find sub1 as a directory with .tf files")
	}
	for _, d := range dirs {
		if strings.HasSuffix(d, "sub2") {
			t.Error("Did not expect sub2 to be found as it has no .tf files")
		}
	}
}

func Test_printStatus(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	printStatus("👍", "Test message")
	w.Close()
	os.Stdout = old
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])
	if !strings.Contains(output, "Test message") {
		t.Error("Expected output to contain status message")
	}
}

func Test_walkDirs_ErrorsAndHidden(t *testing.T) {
	var dirs []string
	err := walkDirs("/nonexistent/path", &dirs)
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}

	tempDir, err2 := os.MkdirTemp("", "walkdirs_test")
	if err2 != nil {
		t.Fatalf("Failed to create temp dir: %v", err2)
	}
	defer os.RemoveAll(tempDir)

	os.Mkdir(filepath.Join(tempDir, ".hidden"), 0755)
	os.Mkdir(filepath.Join(tempDir, ".terraform"), 0755)
	os.Mkdir(filepath.Join(tempDir, "visible"), 0755)
	os.WriteFile(filepath.Join(tempDir, "visible", "main.tf"), []byte("terraform {}"), 0644)
	os.WriteFile(filepath.Join(tempDir, ".hidden", "main.tf"), []byte("terraform {}"), 0644)
	os.WriteFile(filepath.Join(tempDir, ".terraform", "main.tf"), []byte("terraform {}"), 0644)

	var foundDirs []string
	err = walkDirs(tempDir, &foundDirs)
	if err != nil {
		t.Fatalf("walkDirs failed: %v", err)
	}
	foundVisible := false
	foundHidden := false
	foundTerraform := false
	for _, d := range foundDirs {
		if strings.HasSuffix(d, "visible") {
			foundVisible = true
		}
		if strings.HasSuffix(d, ".hidden") {
			foundHidden = true
		}
		if strings.HasSuffix(d, ".terraform") {
			foundTerraform = true
		}
	}
	if !foundVisible {
		t.Error("Expected to find visible directory with .tf files")
	}
	if foundHidden {
		t.Error("Did not expect to find .hidden directory")
	}
	if foundTerraform {
		t.Error("Did not expect to find .terraform directory")
	}
}

func TestRunTofuValidateCLI_AllBranches(t *testing.T) {
	type mockArgs struct {
		checkInstalled bool
		getwdErr       error
		dirs           []string
		runCmdErr      error
		runCmdOut      string
		runValidateErr error
		runValidateOut string
	}
	cases := []struct {
		name     string
		args     mockArgs
		wantErr  bool
		wantExit int
	}{
		{"not installed", mockArgs{checkInstalled: false}, true, 1},
		{"getwd error", mockArgs{checkInstalled: true, getwdErr: fmt.Errorf("fail")}, true, 1},
		{"no tf dirs", mockArgs{checkInstalled: true, dirs: []string{}}, false, 0},
		{"init error", mockArgs{checkInstalled: true, dirs: []string{"/mock"}, runCmdErr: fmt.Errorf("fail"), runCmdOut: "init fail"}, true, 1},
		{"validate error", mockArgs{checkInstalled: true, dirs: []string{"/mock"}, runValidateErr: fmt.Errorf("fail"), runValidateOut: "validate fail"}, true, 1},
		{"all ok", mockArgs{checkInstalled: true, dirs: []string{"/mock"}}, false, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			exited := 0
			checkInstalled := func() bool { return tc.args.checkInstalled }
			getwd := func() (string, error) {
				if tc.args.getwdErr != nil {
					return "", tc.args.getwdErr
				}
				return "/mockroot", nil
			}
			findDirs := func(root string) []string { return tc.args.dirs }
			runCmd := func(dir string, args []string) (string, error) {
				return tc.args.runCmdOut, tc.args.runCmdErr
			}
			runValidate := func(dir string, args []string) (string, error) {
				return tc.args.runValidateOut, tc.args.runValidateErr
			}
			statusMsgs := []string{}
			printStatus := func(emoji, msg string) { statusMsgs = append(statusMsgs, emoji+":"+msg) }
			exit := func(code int) { exited = code }
			err := RunTofuValidateCLI([]string{}, checkInstalled, getwd, findDirs, runCmd, runValidate, printStatus, exit)
			if tc.wantErr && err == nil {
				t.Errorf("Expected error for case %q, got nil", tc.name)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("Did not expect error for case %q, got: %v", tc.name, err)
			}
			if exited != tc.wantExit {
				t.Errorf("Expected exit code %d for case %q, got %d", tc.wantExit, tc.name, exited)
			}
		})
	}

	t.Run("relPath rewriting branch", func(t *testing.T) {
		exited := 0
		checkInstalled := func() bool { return true }
		getwd := func() (string, error) { return "/mockroot", nil }
		findDirs := func(root string) []string { return []string{"/mockroot/subdir"} }
		runCmd := func(dir string, args []string) (string, error) {
			return "init ok", nil
		}
		runValidate := func(dir string, args []string) (string, error) {
			return "validate ok", nil
		}
		var statusMsgs []string
		printStatus := func(emoji, msg string) { statusMsgs = append(statusMsgs, emoji+":"+msg) }
		exit := func(code int) { exited = code }
		err := RunTofuValidateCLI([]string{}, checkInstalled, getwd, findDirs, runCmd, runValidate, printStatus, exit)
		if err != nil {
			t.Errorf("Did not expect error for relPath rewriting branch, got: %v", err)
		}
		if exited != 0 {
			t.Errorf("Expected exit code 0 for relPath rewriting branch, got %d", exited)
		}
		foundRewrite := false
		for _, msg := range statusMsgs {
			if strings.Contains(msg, "subdir") {
				foundRewrite = true
			}
		}
		if !foundRewrite {
			t.Error("Expected relPath rewriting logic to be exercised and subdir to appear in status message")
		}
	})

	t.Run("multi-error summary branch", func(t *testing.T) {
		exited := 0
		checkInstalled := func() bool { return true }
		getwd := func() (string, error) { return "/mockroot", nil }
		findDirs := func(root string) []string { return []string{"/mock1", "/mock2"} }
		runCmd := func(dir string, args []string) (string, error) {
			if dir == "/mock1" {
				return "init fail", fmt.Errorf("fail")
			}
			return "init ok", nil
		}
		runValidate := func(dir string, args []string) (string, error) {
			if dir == "/mock2" {
				return "validate fail", fmt.Errorf("fail")
			}
			return "validate ok", nil
		}
		statusMsgs := []string{}
		printStatus := func(emoji, msg string) { statusMsgs = append(statusMsgs, emoji+":"+msg) }
		exit := func(code int) { exited = code }
		err := RunTofuValidateCLI([]string{}, checkInstalled, getwd, findDirs, runCmd, runValidate, printStatus, exit)
		if err == nil {
			t.Error("Expected error for multi-error summary branch, got nil")
		}
		if exited != 1 {
			t.Errorf("Expected exit code 1 for multi-error summary branch, got %d", exited)
		}
	})
}

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
	if _, err := exec.LookPath("tofu"); err != nil {
		t.Skip("Skipping test as tofu is not installed")
	}
	tempDir, err := os.MkdirTemp("", "tofu_validate_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	validDir := filepath.Join(tempDir, "valid")
	if err := os.Mkdir(validDir, 0755); err != nil {
		t.Fatalf("Failed to create valid config directory: %v", err)
	}
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
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(originalDir)
	if err := os.Chdir(validDir); err != nil {
		t.Fatalf("Failed to change directory to %s: %v", validDir, err)
	}
	initCmd := exec.Command("tofu", "init")
	if initOutput, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to initialize tofu: %v, output: %s", err, initOutput)
	}
	validateCmd := exec.Command("tofu", "validate")
	if validateOutput, err := validateCmd.CombinedOutput(); err != nil {
		t.Fatalf("Expected valid config to pass validation, but it failed: %v, output: %s", err, validateOutput)
	}
}

// TestInvalidOpenTofuConfig tests that an invalid config fails validation
func TestInvalidOpenTofuConfig(t *testing.T) {
	if _, err := exec.LookPath("tofu"); err != nil {
		t.Skip("Skipping test as tofu is not installed")
	}
	tempDir, err := os.MkdirTemp("", "tofu_validate_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	invalidDir := filepath.Join(tempDir, "invalid")
	if err := os.Mkdir(invalidDir, 0755); err != nil {
		t.Fatalf("Failed to create invalid config directory: %v", err)
	}
	invalidContent := `tofu {
  required_version = ">= 1.0.0"
  # Missing closing brace intentionally`
	invalidFilePath := filepath.Join(invalidDir, "main.tf")
	if err := os.WriteFile(invalidFilePath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to write invalid file: %v", err)
	}
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(originalDir)
	if err := os.Chdir(invalidDir); err != nil {
		t.Fatalf("Failed to change directory to %s: %v", invalidDir, err)
	}
	validateCmd := exec.Command("tofu", "validate")
	output, err := validateCmd.CombinedOutput()
	if err == nil {
		t.Fatalf("Expected invalid config to fail validation, but it passed. Output: %s", output)
	}
}
