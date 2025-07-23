package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"pre-commit-hooks/internal/outputs"
	"pre-commit-hooks/internal/testutil"
	tofu_fmt "pre-commit-hooks/internal/tofu_fmt"
	"testing"
)

func TestRunTofuFmtCLI_AllBranches(t *testing.T) {
	type mockArgs struct {
		checkInstalled bool
		getwdErr       error
		runFmtErr      error
		runFmtOut      string
		formatErr      error
	}
	cases := []struct {
		name    string
		args    mockArgs
		wantErr bool
	}{
		{"not installed", mockArgs{checkInstalled: false}, true},
		{"getwd error", mockArgs{checkInstalled: true, getwdErr: fmt.Errorf("fail")}, true},
		{"all formatted", mockArgs{checkInstalled: true}, false},
		{"unformatted, format ok", mockArgs{checkInstalled: true, runFmtErr: fmt.Errorf("unformatted"), runFmtOut: "needs format"}, false},
		{"unformatted, format fails", mockArgs{checkInstalled: true, runFmtErr: fmt.Errorf("unformatted"), runFmtOut: "needs format", formatErr: fmt.Errorf("format fail")}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			getwd := func() (string, error) {
				if tc.args.getwdErr != nil {
					return "", tc.args.getwdErr
				}
				return "/tmp/mock", nil
			}
			runFmt := func(dir string, args []string) (string, error) {
				return tc.args.runFmtOut, tc.args.runFmtErr
			}
			format := func(dir string, args []string) error {
				return tc.args.formatErr
			}
			// Patch tofu_fmt.CheckOpenTofuInstalled for this test
			origCheck := tofu_fmt.CheckOpenTofuInstalled
			tofu_fmt.CheckOpenTofuInstalled = func() bool { return tc.args.checkInstalled }
			defer func() { tofu_fmt.CheckOpenTofuInstalled = origCheck }()
			err := RunTofuFmtCLI([]string{}, getwd, runFmt, format)
			if tc.wantErr && err == nil {
				t.Errorf("Expected error for case %q, got nil", tc.name)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("Did not expect error for case %q, got: %v", tc.name, err)
			}
		})
	}
}

func TestRunTofuFmtCLI_NotInstalled(t *testing.T) {
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", oldPath)
	err := RunTofuFmtCLI([]string{}, os.Getwd, tofu_fmt.RunTofuFmt, tofu_fmt.FormatFiles)
	if err == nil {
		t.Error("Expected error when OpenTofu is not installed")
	}
}

func TestRunTofuFmtCLI_BadDir(t *testing.T) {
	// Simulate error getting working directory
	err := RunTofuFmtCLI([]string{}, func() (string, error) { return "", fmt.Errorf("fail") }, tofu_fmt.RunTofuFmt, tofu_fmt.FormatFiles)
	if err == nil {
		t.Error("Expected error when failing to get working directory")
	}
}

func TestPrintStatus(t *testing.T) {
	cases := []struct {
		emoji string
		msg   string
		color string
		want  string
	}{
		{outputs.ThumbsUp, "All good", outputs.Green, outputs.ThumbsUp + " " + outputs.Green + "All good" + outputs.Reset + "\n"},
		{outputs.Warning, "Warning!", outputs.Yellow, outputs.Warning + " " + outputs.Yellow + "Warning!" + outputs.Reset + "\n"},
		{outputs.Error, "Error!", outputs.Red, outputs.Error + " " + outputs.Red + "Error!" + outputs.Reset + "\n"},
	}
	for _, c := range cases {
		buf := &bytes.Buffer{}
		fmt.Fprintf(buf, "%s\n", outputs.EmojiColorText(c.emoji, c.msg, c.color))
		got := buf.String()
		if got != c.want {
			t.Errorf("printStatus(%q, %q): got %q, want %q", c.emoji, c.msg, got, c.want)
		}
	}
}

func TestMain_ErrorHandling(t *testing.T) {
	// Simulate tofu not installed
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", oldPath)
	if testutil.CheckOpenTofuInstalled() {
		t.Error("Expected CheckOpenTofuInstalled to be false when PATH is empty")
	}
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

func TestTofuFmt(t *testing.T) {
	// Skip test if tofu is not installed
	if _, err := exec.LookPath("tofu"); err != nil {
		t.Skip("Skipping test as tofu is not installed")
	}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "tofu_fmt_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create formatted and unformatted files
	formattedContent := `variable "example" {}`
	unformattedContent := `variable   "example"   {}` // Extra spaces

	formattedFilePath := filepath.Join(tempDir, "formatted.tf")
	unformattedFilePath := filepath.Join(tempDir, "unformatted.tf")

	if err := os.WriteFile(formattedFilePath, []byte(formattedContent), 0644); err != nil {
		t.Fatalf("Failed to write formatted file: %v", err)
	}
	if err := os.WriteFile(unformattedFilePath, []byte(unformattedContent), 0644); err != nil {
		t.Fatalf("Failed to write unformatted file: %v", err)
	}

	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory to %s: %v", tempDir, err)
	}

	// Run tofu fmt
	fmtCmd := exec.Command("tofu", "fmt")
	if output, err := fmtCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to run tofu fmt: %v, output: %s", err, output)
	}

	// Verify unformatted file is now formatted
	result, err := os.ReadFile(unformattedFilePath)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", unformattedFilePath, err)
	}
	if string(result) != formattedContent {
		t.Errorf("File was not formatted correctly. Got: %q, Want: %q", string(result), formattedContent)
	}
}
