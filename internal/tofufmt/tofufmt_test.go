package tofufmt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pre-commit-hooks/internal/testutil"
)

func TestCheckOpenTofuInstalled(t *testing.T) {
	testutil.SkipIfTofuNotInstalled(t)
	got := CheckOpenTofuInstalled()
	if !got {
		t.Error("Expected CheckOpenTofuInstalled to return true when tofu is installed, but got false.")
	}
}

func TestRunTofuFmt_MultiFileAndNested(t *testing.T) {
	testutil.SkipIfTofuNotInstalled(t)
	tempDir, cleanup := testutil.CreateTempDir(t, "fmt_test_multi")
	defer cleanup()

	// Table of files to create: path, content, expected after format
	files := []struct {
		relPath string
		content string
		want    string
	}{
		{"main.tf", "variable   \"foo\"   {}", "variable \"foo\" {}"},
		{"formatted.tf", "variable \"bar\" {}", "variable \"bar\" {}"},
		{"subdir/nested.tf", "variable   \"baz\"   {}", "variable \"baz\" {}"},
		{"subdir/ignore.txt", "not terraform", "not terraform"},
	}

	// Create files
	for _, f := range files {
		fullPath := filepath.Join(tempDir, f.relPath)
		dirPath := filepath.Dir(fullPath)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dirPath, err)
		}
		if err := os.WriteFile(fullPath, []byte(f.content), 0644); err != nil {
			t.Fatalf("Failed to write test file %s: %v", fullPath, err)
		}
	}

	// Run tofu fmt check
	output, err := RunTofuFmt(tempDir, nil)
	if err == nil {
		t.Fatalf("Expected tofu fmt to report unformatted files, but got no error. Output: %s", output)
	}

	// Format all files
	if err := FormatFiles(tempDir, nil); err != nil {
		t.Fatalf("Failed to format files: %v", err)
	}

	// Check that all .tf files are formatted
	for _, f := range files {
		if !strings.HasSuffix(f.relPath, ".tf") {
			continue // skip non-tf files
		}
		fullPath := filepath.Join(tempDir, f.relPath)
		result, err := os.ReadFile(fullPath)
		if err != nil {
			t.Fatalf("Failed to read file %s: %v", fullPath, err)
		}
		if string(result) != f.want {
			t.Errorf("File %s was not formatted correctly. Got: %q, Want: %q", f.relPath, string(result), f.want)
		}
	}
}

func TestRunTofuFmt_UnformattedFile(t *testing.T) {
	testutil.SkipIfTofuNotInstalled(t)
	tempDir, cleanup := testutil.CreateTempDir(t, "fmt_test_unformatted")
	defer cleanup()

	// Create an unformatted .tf file
	filePath := filepath.Join(tempDir, "main.tf")
	unformatted := "variable   \"foo\"   {}"
	formatted := "variable \"foo\" {}"
	if err := os.WriteFile(filePath, []byte(unformatted), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	output, err := RunTofuFmt(tempDir, nil)
	if err == nil {
		t.Fatalf("Expected tofu fmt to report unformatted file, but got no error. Output: %s", output)
	}
	// Now format the file
	if err := FormatFiles(tempDir, nil); err != nil {
		t.Fatalf("Failed to format file: %v", err)
	}
	// Check that file is now formatted
	result, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(result) != formatted {
		t.Errorf("File was not formatted correctly. Got: %q, Want: %q", string(result), formatted)
	}
}
