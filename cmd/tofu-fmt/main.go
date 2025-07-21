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

func runTofuCmdCapture(baseArgs, extraArgs []string) ([]byte, error) {
	args := append(baseArgs, extraArgs...)
	cmd := exec.Command("tofu", args...)
	return cmd.CombinedOutput()
}

func runTofuCmdStream(baseArgs, extraArgs []string) error {
	args := append(baseArgs, extraArgs...)
	cmd := exec.Command("tofu", args...)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
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
	<-done
	<-done
	return cmd.Wait()
}

// scanAndIndent prints each line from r with indentation
func scanAndIndent(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fmt.Printf("    %s\n", scanner.Text())
	}
}

func main() {
	if !checkOpenTofuInstalled() {
		fmt.Println("OpenTofu is not installed or not in PATH.")
		os.Exit(1)
	}
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		os.Exit(1)
	}
	dirName := wd[strings.LastIndex(wd, string(os.PathSeparator))+1:]
	fmt.Printf("Running tofu fmt recursively in directory: %s\n", dirName)

	extraArgs := os.Args[1:]

	printStatus(outputs.Running, "Running tofu fmt...")

	output, err := runTofuCmdCapture([]string{"fmt", "-check", "-recursive", "--diff"}, extraArgs)
	fmt.Println()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 3 {
			fmt.Println(outputs.EmojiColorText(outputs.Warning, "Found unformatted OpenTofu files:", outputs.Yellow))
			scanAndIndent(strings.NewReader(string(output)))
			printStatus(outputs.Running, "Formatting files with tofu fmt...")
			fmtErr := runTofuCmdStream([]string{"fmt", "-recursive"}, extraArgs)
			fmt.Println()
			if fmtErr != nil {
				printStatus(outputs.Error, "Error running tofu fmt:")
				fmt.Println(fmtErr)
				os.Exit(1)
			}
			printStatus(outputs.ThumbsUp, "Files formatted successfully with tofu fmt.")
			fmt.Println()
		} else {
			printStatus(outputs.Error, "Error running tofu fmt:")
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		printStatus(outputs.ThumbsUp, "All OpenTofu files are formatted.")
		fmt.Println()
	}
}

// printStatus prints a colored emoji status message
func printStatus(emoji, msg string) {
	fmt.Println(outputs.EmojiColorText(emoji, msg, outputs.Green))
}
