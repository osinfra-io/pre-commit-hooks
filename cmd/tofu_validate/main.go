package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"pre-commit-hooks/internal/outputs"
	tofu_validate "pre-commit-hooks/internal/tofu_validate"
)

func main() {
	err := RunTofuValidateCLI(
		os.Args[1:],
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
		printStatus(outputs.Running, fmt.Sprintf("Running tofu init in: %s...", relPath))
		initCmd := []string{"init", "-input=false", "--backend=false"}
		cmdArgs := append(initCmd, extraArgs...)
		out, err := runCmd(dir, cmdArgs)
		fmt.Print(out)
		fmt.Println()
		if err != nil {
			errorMessages = append(errorMessages, tofuError{"init", relPath, out})
			continue
		}

		printStatus(outputs.Running, fmt.Sprintf("Running tofu validate in: %s...", relPath))
		out, err = runValidate(dir, extraArgs)
		fmt.Print(out)
		fmt.Println()
		if err != nil {
			errorMessages = append(errorMessages, tofuError{"validate", relPath, out})
			continue
		}
	}

	if len(errorMessages) > 0 {
		fmt.Println(outputs.EmojiColorText("⚠️", "Validation Summary:", outputs.Yellow))
		fmt.Println()
		for _, msg := range errorMessages {
			fmt.Printf(outputs.EmojiColorText(outputs.Error, "OpenTofu %s failed in: %s\n%s\n", outputs.Red), msg.step, msg.relPath, msg.output)
		}
		exit(1)
		return fmt.Errorf("Some directories failed validation")
	} else {
		printStatus(outputs.ThumbsUp, "OpenTofu validate completed successfully for all directories.")
		fmt.Println()
	}
	return nil
}

// runCmdInDir runs a command in the specified directory, returns all output and error
func runCmdInDir(dir string, args []string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
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
			walkDirs(dir+string(os.PathSeparator)+name, dirs)
		} else if strings.HasSuffix(name, ".tf") {
			hasTf = true
		}
	}
	if hasTf {
		*dirs = append(*dirs, dir)
	}
	return nil
}

// printStatus prints a colored emoji status message
func printStatus(emoji, msg string) {
	fmt.Println(outputs.EmojiColorText(emoji, msg, outputs.Green))
}
