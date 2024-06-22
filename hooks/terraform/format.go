package main

import (
	"fmt"
	"github.com/osinfra-io/pre-commit-hooks/hooks/helpers"
	"os"
	"os/exec"
	"strings"
)

func main() {
	fmt.Println(outputs.CyanRunning("Running terraform fmt..."))

	// Find unformatted Terraform files
	cmd := exec.Command("terraform", "fmt", "-check", "-recursive")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if the error is an ExitError and the exit code is 3
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 3 {
			fmt.Println(outputs.YellowWarning("Found unformatted Terraform files:"))
		} else {
			fmt.Printf(outputs.RedError("Error running terraform fmt: %v\n"), err)
			os.Exit(1)
		}
	}

	unformattedFiles := strings.TrimSpace(string(output))
	if unformattedFiles != "" {
		// Split the unformattedFiles string into a slice of file names
		fileNames := strings.Split(unformattedFiles, "\n")
		for _, file := range fileNames {
			fmt.Println("  " + outputs.YellowDiamond(file))
		}

		fmt.Println(outputs.CyanWorking("Formatting files with terraform fmt..."))
		cmd := exec.Command("terraform", "fmt", "-recursive")
		err := cmd.Run()

		if err != nil {
			fmt.Println("Error running terraform fmt:", err)
		} else {
			fmt.Println(outputs.GreenThumbsUp("Files formatted successfully with terraform fmt."))
		}
	} else {
		// This line will now only run if unformattedFiles is empty
		fmt.Println(outputs.GreenThumbsUp("All Terraform files are formatted."))
	}
}
