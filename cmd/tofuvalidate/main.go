package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"pre-commit-hooks/internal/output"
	"pre-commit-hooks/internal/testutil"
)

func main() {
	// Only pass flags (arguments starting with '-') to tofu commands
	extraArgs := []string{}
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-") {
			extraArgs = append(extraArgs, arg)
		}
	}
	err := RunTofuValidateCLI(
		extraArgs,
		testutil.CheckOpenTofuInstalled,
		os.Getwd,
		findDirsWithTfFiles,
		runCmdInDir,
		runTofuValidate,
	)
	if err != nil {
		os.Exit(1)
	}
}

// RunTofuValidateCLI runs the tofu validate CLI logic. Returns error if any step fails.
func RunTofuValidateCLI(
	extraArgs []string,
	checkInstalled func() bool,
	getwd func() (string, error),
	findDirs func(string) []string,
	runCmd func(string, []string) (string, error),
	runValidate func(string, []string) (string, error),
) error {
	if !checkInstalled() {
		fmt.Println("OpenTofu is not installed or not in PATH.")
		return fmt.Errorf("OpenTofu not installed")
	}

	rootDir, err := getwd()
	if err != nil {
		fmt.Println("Could not get working directory.")
		return err
	}

	dirsWithTf := findDirs(rootDir)
	if len(dirsWithTf) == 0 {
		fmt.Println("No directories with Terraform files found.")
		return nil
	}

	var errorMessages []output.TofuMessage
	var warningMessages []output.TofuMessage
	baseDir := filepath.Base(rootDir)
	for _, dir := range dirsWithTf {
		relPath, err := filepath.Rel(rootDir, dir)
		if err != nil {
			relPath = dir // fallback to absolute path
		}
		var fullPath string
		if relPath == "." {
			fullPath = baseDir
		} else if strings.HasPrefix(relPath, "..") {
			// If path is outside rootDir, use just the dir name
			fullPath = filepath.Base(dir)
		} else {
			fullPath = baseDir + "/" + relPath
		}
		initCmd := []string{"init", "-input=false", "--backend=false"}
		cmdArgs := append(initCmd, extraArgs...)
		out, err := runCmd(dir, cmdArgs)
		if err != nil {
			errorMessages = append(errorMessages, output.TofuMessage{Step: "init", RelPath: fullPath, Output: out})
			continue
		}

		out, err = runValidate(dir, extraArgs)
		// Always check for warnings in validate output
		if hasWarning(out) {
			warningMessages = append(warningMessages, output.TofuMessage{Step: "validate", RelPath: fullPath, Output: out})
		}
		if err != nil {
			errorMessages = append(errorMessages, output.TofuMessage{Step: "validate", RelPath: fullPath, Output: out})
			continue
		}
	}

	if len(warningMessages) > 0 {
		output.PrintWarningSummary(warningMessages)
	}

	if len(errorMessages) > 0 {
		output.PrintErrorSummary(errorMessages)
		return fmt.Errorf("validation failed")
	}

	return nil
}

func runTofuValidate(dir string, extraArgs []string) (string, error) {
	args := append([]string{"validate"}, extraArgs...)
	cmd := exec.Command("tofu", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// runCmdInDir runs a command in the specified directory, returns all output and error
func runCmdInDir(dir string, args []string) (string, error) {
	cmd := exec.Command("tofu", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// findDirsWithTfFiles recursively finds directories containing .tf files
func findDirsWithTfFiles(root string) []string {
	dirs, err := walkDirs(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Error scanning directories: %v\n", err)
	}
	return dirs
}

func walkDirs(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var dirs []string
	var errs []error
	hasTfOrTofu := false
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			if strings.HasPrefix(name, ".") {
				continue
			}
			path := filepath.Join(dir, name)
			subDirs, err := walkDirs(path)
			if err != nil {
				errs = append(errs, err)
			}
			dirs = append(dirs, subDirs...)
		} else if strings.HasSuffix(name, ".tf") || strings.HasSuffix(name, ".tofu") {
			hasTfOrTofu = true
		}
	}
	if hasTfOrTofu {
		dirs = append(dirs, dir)
	}
	return dirs, errors.Join(errs...)
}

// hasWarning checks if output contains a warning message
// Uses pattern matching to avoid false positives from filenames or unrelated text
func hasWarning(output string) bool {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "warning:") ||
			strings.HasPrefix(lower, "│ warning:") ||
			(strings.HasPrefix(lower, "╷") && strings.Contains(lower, "warning")) {
			return true
		}
	}
	return false
}
