package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"pre-commit-hooks/internal/output"
	tofu_validate "pre-commit-hooks/internal/tofuvalidate"
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
		tofu_validate.CheckOpenTofuInstalled,
		os.Getwd,
		findDirsWithTfFiles,
		runCmdInDir,
		tofu_validate.RunTofuValidate,
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

	type tofuError struct {
		step    string
		relPath string
		output  string
	}
	var errorMessages []tofuError
	for _, dir := range dirsWithTf {
		relPath := dir
		if strings.HasPrefix(dir, rootDir) {
			relPath = strings.TrimPrefix(dir, rootDir)
			relPath = strings.TrimPrefix(relPath, string(os.PathSeparator))
			relPath = rootDir[strings.LastIndex(rootDir, string(os.PathSeparator))+1:] + string(os.PathSeparator) + relPath
		}
		printStatus(output.Running, fmt.Sprintf("Running tofu init in: %s...", relPath))
		initCmd := []string{"init", "-input=false", "--backend=false"}
		cmdArgs := append(initCmd, extraArgs...)
		out, err := runCmd(dir, cmdArgs)
		printIndentedOutput(out, true)
		if err != nil {
			errorMessages = append(errorMessages, tofuError{"init", relPath, out})
			continue
		}

		printStatus(output.Running, fmt.Sprintf("Running tofu validate in: %s...", relPath))
		out, err = runValidate(dir, extraArgs)
		printIndentedOutput(out, true)
		if err != nil {
			errorMessages = append(errorMessages, tofuError{"validate", relPath, out})
			continue
		}
	}

	if len(errorMessages) > 0 {
		fmt.Println(output.EmojiColorText("⚠️", "Validation Summary:", output.Yellow))
		fmt.Println()
		for _, msg := range errorMessages {
			fmt.Printf(output.EmojiColorText(output.Error, "OpenTofu %s failed in: %s\n", output.Red), msg.step, msg.relPath)
			printIndentedOutput(msg.output, false)
		}
		exit(1)
		return fmt.Errorf("validation failed")
	} else {
		printStatus(output.ThumbsUp, "OpenTofu validate completed successfully for all directories.")
		fmt.Println()
	}
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
	hasTf := false
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
		} else if strings.HasSuffix(name, ".tf") {
			hasTf = true
		}
	}
	if hasTf {
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
