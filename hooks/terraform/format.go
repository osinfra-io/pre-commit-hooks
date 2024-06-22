package main

import (
	"fmt"
	outputs "github.com/osinfra-io/pre-commit-hooks/hooks/helpers"
	"os"
	"os/exec"
	"strings"
)

func main() {
	fmt.Println(outputs.EmojiColorText(outputs.Running, "Running terraform fmt...", outputs.Cyan))

	// Find unformatted Terraform files
	cmd := exec.Command("terraform", "fmt", "-check", "-recursive")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if the error is an ExitError and the exit code is 3
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 3 {
			fmt.Println(outputs.EmojiColorText(outputs.Warning, "Found unformatted Terraform files:", outputs.Yellow))
		} else {
			fmt.Printf(outputs.EmojiColorText(outputs.Error, "Error running terraform fmt: %v\n", outputs.Red), err)
			os.Exit(1)
		}
	}

	unformattedFiles := strings.TrimSpace(string(output))
	if unformattedFiles != "" {
		// Split the unformattedFiles string into a slice of file names
		fileNames := strings.Split(unformattedFiles, "\n")
		for _, file := range fileNames {
			fmt.Println("  " + outputs.EmojiColorText(outputs.Diamond, (file), outputs.Yellow))
		}

		fmt.Println(outputs.EmojiColorText(outputs.Working, "Formatting files with terraform fmt...", outputs.Cyan))
		cmd := exec.Command("terraform", "fmt", "-recursive")
		err := cmd.Run()

		if err != nil {
			fmt.Println(outputs.EmojiColorText(outputs.Error, "Error running terraform fmt:", outputs.Red), err)
		} else {
			fmt.Println(outputs.EmojiColorText(outputs.ThumbsUp, "Files formatted successfully with terraform fmt.", outputs.Green))
		}
	} else {
		// This line will now only run if unformattedFiles is empty
		fmt.Println(outputs.EmojiColorText(outputs.ThumbsUp, "All Terraform files are formatted.", outputs.Green))
	}
}
