package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"pre-commit-hooks/internal/outputs"
)

func checkOpenTofuInstalled() bool {
	_, err := exec.LookPath("tofu")
	return err == nil
}

func main() {
	if !checkOpenTofuInstalled() {
		fmt.Println("OpenTofu is not installed or not in PATH.")
		os.Exit(1)
	}

	// Only pass flags (arguments starting with '-') to tofu commands
	extraArgs := []string{}
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-") {
			extraArgs = append(extraArgs, arg)
		}
	}

	rootDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Could not get working directory.")
		os.Exit(1)
	}

	dirsWithTf := findDirsWithTfFiles(rootDir)
	if len(dirsWithTf) == 0 {
		fmt.Println("No directories with Terraform files found.")
		os.Exit(0)
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
		out, err := runTofuCmdInDirWithOutput(dir, []string{"init", "-input=false", "--backend=false"}, extraArgs)
		fmt.Print(out)
		fmt.Println()
		if err != nil {
			errorMessages = append(errorMessages, tofuError{"init", relPath, out})
			continue
		}

		printStatus(outputs.Running, fmt.Sprintf("Running tofu validate in: %s...", relPath))
		out, err = runTofuCmdInDirWithOutput(dir, []string{"validate"}, extraArgs)
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
		os.Exit(1)
	} else {
		printStatus(outputs.ThumbsUp, "OpenTofu validate completed successfully for all directories.")
		fmt.Println()
	}
}

// runTofuCmdInDirWithOutput runs a tofu command in the specified directory, returns all output and error
func runTofuCmdInDirWithOutput(dir string, baseArgs, extraArgs []string) (string, error) {
	args := append(baseArgs, extraArgs...)
	cmd := exec.Command("tofu", args...)
	cmd.Dir = dir
	var combinedOut strings.Builder
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}
	if err := cmd.Start(); err != nil {
		return "", err
	}
	done := make(chan struct{})
	go func() {
		scanAndCollect(stdoutPipe, &combinedOut)
		done <- struct{}{}
	}()
	go func() {
		scanAndCollect(stderrPipe, &combinedOut)
		done <- struct{}{}
	}()
	<-done
	<-done
	err = cmd.Wait()
	return combinedOut.String(), err
}

// scanAndCollect collects each line from r into out
func scanAndCollect(r io.Reader, out *strings.Builder) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		out.WriteString("    " + line + "\n")
	}
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
