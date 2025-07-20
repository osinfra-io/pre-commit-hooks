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

	fmt.Println(outputs.EmojiColorText(outputs.Running, "Running tofu fmt...", outputs.Green))

	// Collect extra args from .pre-commit-config.yaml via os.Args[1:]
	extraArgs := os.Args[1:]
	checkArgs := []string{"fmt", "-check", "-recursive"}
	checkArgs = append(checkArgs, extraArgs...)
	cmd := exec.Command("tofu", checkArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if the error is an ExitError and the exit code is 3
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 3 {
			fmt.Println(outputs.EmojiColorText(outputs.Warning, "Found unformatted OpenTofu files:", outputs.Yellow))
		} else {
			fmt.Printf(outputs.EmojiColorText(outputs.Error, "Error running tofu fmt: %v\n", outputs.Red), err)
			os.Exit(1)
		}
	}

	unformattedFiles := strings.TrimSpace(string(output))
	if unformattedFiles != "" {
		// Split the unformattedFiles string into a slice of file names
		fileNames := strings.Split(unformattedFiles, "\n")
		for _, file := range fileNames {
			fmt.Println(outputs.Warning, "  - "+file)
		}

		fmt.Println(outputs.EmojiColorText(outputs.Working, "Formatting files with tofu fmt...", outputs.Green))
		fmtArgs := []string{"fmt", "-recursive"}
		fmtArgs = append(fmtArgs, extraArgs...)
		cmd := exec.Command("tofu", fmtArgs...)
		err := cmd.Run()

		if err != nil {
			fmt.Println(outputs.EmojiColorText(outputs.Error, "Error running tofu fmt:", outputs.Red), err)
		} else {
			fmt.Println(outputs.EmojiColorText(outputs.ThumbsUp, "Files formatted successfully with tofu fmt.", outputs.Green))
		}
	} else {
		// This line will now only run if unformattedFiles is empty
		fmt.Println(outputs.EmojiColorText(outputs.ThumbsUp, "All OpenTofu files are formatted.", outputs.Green))
	}
}
