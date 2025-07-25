package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"pre-commit-hooks/internal/output"
	tofuvalidate "pre-commit-hooks/internal/tofuvalidate"
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
		tofuvalidate.CheckOpenTofuInstalled,
		os.Getwd,
		findDirsWithTfFiles,
		runCmdInDir,
		tofuvalidate.RunTofuValidate,
		printStatus,
		os.Exit,
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
	printStatus func(string, string),
	exit func(int),
) error {
	if !checkInstalled() {
		fmt.Println("OpenTofu is not installed or not in PATH.")
		exit(1)
		return fmt.Errorf("OpenTofu not installed")
	}

	rootDir, err := getwd()
	if err != nil {
		fmt.Println("Could not get working directory.")
		exit(1)
		return err
	}

	dirsWithTf := findDirs(rootDir)
	if len(dirsWithTf) == 0 {
		fmt.Println("No directories with Terraform files found.")
		exit(0)
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
		fullPath := relPath
		if relPath == "." {
			fullPath = baseDir
		} else {
			fullPath = baseDir + "/" + relPath
		}
		printStatus(output.Running, fmt.Sprintf("Running tofu init in: %s...", fullPath))
		initCmd := []string{"init", "-input=false", "--backend=false"}
		cmdArgs := append(initCmd, extraArgs...)
		out, err := runCmd(dir, cmdArgs)
		printIndentedOutput(out, true)
		// Always check for warnings in init output
		if strings.Contains(strings.ToLower(out), "warning") {
			warningMessages = append(warningMessages, output.TofuMessage{Step: "init", RelPath: fullPath, Output: out})
		}
		if err != nil {
			errorMessages = append(errorMessages, output.TofuMessage{Step: "init", RelPath: fullPath, Output: out})
			continue
		}

		printStatus(output.Running, fmt.Sprintf("Running tofu validate in: %s...", fullPath))
		out, err = runValidate(dir, extraArgs)
		printIndentedOutput(out, true)
		// Always check for warnings in validate output
		if strings.Contains(strings.ToLower(out), "warning") {
			warningMessages = append(warningMessages, output.TofuMessage{Step: "validate", RelPath: fullPath, Output: out})
		}
		if err != nil {
			// Only treat as error if not a warning (warnings already handled above)
			if !strings.Contains(strings.ToLower(out), "warning") {
				errorMessages = append(errorMessages, output.TofuMessage{Step: "validate", RelPath: fullPath, Output: out})
			}
			continue
		}
	}

	if len(warningMessages) > 0 {
		output.PrintWarningSummary(warningMessages)
	}

	if len(errorMessages) > 0 {
		output.PrintErrorSummary(errorMessages, printIndentedOutput)
		exit(1)
		return fmt.Errorf("validation failed")
	}

	if len(warningMessages) > 0 {
		printStatus(output.ThumbsUp, "OpenTofu validate completed with warnings.")
		fmt.Println()
		exit(0)
		return nil
	}

	printStatus(output.ThumbsUp, "OpenTofu validate completed successfully for all directories.")
	fmt.Println()
	return nil
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
	var dirs []string
	_ = walkDirs(root, &dirs)
	return dirs
}

func walkDirs(dir string, dirs *[]string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	hasTfOrTofu := false
	for _, entry := range entries {
		name := entry.Name()
		// Skip hidden/system folders
		if entry.IsDir() {
			if strings.HasPrefix(name, ".") || name == ".terraform" {
				continue
			}
			path := filepath.Join(dir, name)
			if err := walkDirs(path, dirs); err != nil {
				return err
			}
		} else if strings.HasSuffix(name, ".tf") || strings.HasSuffix(name, ".tofu") {
			hasTfOrTofu = true
		}
	}
	if hasTfOrTofu {
		*dirs = append(*dirs, dir)
	}
	return nil
}

// printIndentedOutput prints each line of output indented for better readability
func printIndentedOutput(output string, addNewline bool) {
	lines := strings.Split(output, "\n")
	lastNonEmpty := -1
	for idx := range lines {
		if strings.TrimSpace(lines[idx]) != "" {
			lastNonEmpty = idx
		}
	}
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			fmt.Printf("    %s\n", line)
		}
	}
	// Only add newline if not already present at the end
	if addNewline && lastNonEmpty != len(lines)-1 {
		fmt.Println()
	}
}

// printStatus prints a colored emoji status message
func printStatus(emoji, msg string) {
	fmt.Println(output.EmojiColorText(emoji, msg, output.Green))
}
