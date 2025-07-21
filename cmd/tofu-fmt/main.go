package main

import (
	"fmt"
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
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		os.Exit(1)
	}
	dirName := wd[strings.LastIndex(wd, string(os.PathSeparator))+1:]
	fmt.Printf("Running tofu fmt recursively in directory: %s\n", dirName)

	extraArgs := os.Args[1:]

	printStatus(outputs.Running, "Running tofu fmt...")

	output, err := runTofuCmd([]string{"fmt", "-check", "-recursive"}, extraArgs)
	unformattedFiles := strings.TrimSpace(string(output))

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 3 && unformattedFiles != "" {
			fmt.Println(outputs.EmojiColorText(outputs.Warning, "Found unformatted OpenTofu files:", outputs.Yellow))
			for _, file := range strings.Split(unformattedFiles, "\n") {
				fmt.Println(outputs.EmojiColorText(outputs.Warning, "  - "+file, outputs.Yellow))
			}
			printStatus(outputs.Running, "Formatting files with tofu fmt...")
			_, fmtErr := runTofuCmd([]string{"fmt", "-recursive"}, extraArgs)
			if fmtErr != nil {
				printStatus(outputs.Error, "Error running tofu fmt:")
				fmt.Println(fmtErr)
				os.Exit(1)
			}
			printStatus(outputs.ThumbsUp, "Files formatted successfully with tofu fmt.")
		} else {
			printStatus(outputs.Error, "Error running tofu fmt:")
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		printStatus(outputs.ThumbsUp, "All OpenTofu files are formatted.")
	}
}

// runTofuCmd runs a tofu command with base and extra args, returns output and error
func runTofuCmd(baseArgs, extraArgs []string) ([]byte, error) {
	args := append(baseArgs, extraArgs...)
	cmd := exec.Command("tofu", args...)
	return cmd.CombinedOutput()
}

// printStatus prints a colored emoji status message
func printStatus(emoji, msg string) {
	fmt.Println(outputs.EmojiColorText(emoji, msg, outputs.Green))
}
