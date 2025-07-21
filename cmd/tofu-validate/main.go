package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

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

	for _, dir := range dirsWithTf {
		relPath := dir
		if strings.HasPrefix(dir, rootDir) {
			relPath = strings.TrimPrefix(dir, rootDir)
			relPath = strings.TrimPrefix(relPath, string(os.PathSeparator))
			relPath = rootDir[strings.LastIndex(rootDir, string(os.PathSeparator))+1:] + string(os.PathSeparator) + relPath
		}
		printStatus(outputs.Running, fmt.Sprintf("Running tofu init in %s...", relPath))
		output, err := runTofuCmdInDir(dir, []string{"init", "-input=false", "--backend=false"}, extraArgs)
		fmt.Println()
		if err != nil {
			printError(fmt.Sprintf("OpenTofu init failed in %s", relPath), err, output)
			os.Exit(1)
		}

		printStatus(outputs.Running, fmt.Sprintf("Running tofu validate in %s...", relPath))
		output, err = runTofuCmdInDir(dir, []string{"validate"}, extraArgs)
		fmt.Println()
		if err != nil {
			printError(fmt.Sprintf("OpenTofu validate failed in %s", relPath), err, output)
			os.Exit(1)
		}
	}

	printStatus(outputs.ThumbsUp, "OpenTofu validate completed successfully for all directories.")
}

// runTofuCmdInDir runs a tofu command in the specified directory
func runTofuCmdInDir(dir string, baseArgs, extraArgs []string) ([]byte, error) {
	args := append(baseArgs, extraArgs...)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "tofu", args...)
	cmd.Dir = dir

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	done := make(chan struct{})

	go func() {
		scanAndIndent(stdoutPipe)
		done <- struct{}{}
	}()
	go func() {
		scanAndIndent(stderrPipe)
		done <- struct{}{}
	}()

	// Wait for both pipes to finish
	<-done
	<-done

	err = cmd.Wait()
	return nil, err
}

// scanAndIndent prints each line from r with indentation
func scanAndIndent(r io.Reader) {

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fmt.Printf("    %s\n", scanner.Text())
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
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules" || name == ".terraform" || name == ".git" {
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

// printError prints a colored error message with output
func printError(prefix string, err error, output []byte) {
	fmt.Printf(outputs.EmojiColorText(outputs.Error, "%s: %v\n%s", outputs.Red), prefix, err, output)
}
