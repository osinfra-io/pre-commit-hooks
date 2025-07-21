package main

import (
	"fmt"
	"os"
	"os/exec"

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

       cwd, err := os.Getwd()
       if err != nil {
	       cwd = "(unknown directory)"
       }
       extraArgs := os.Args[1:]

       printStatus(outputs.Running, fmt.Sprintf("Running tofu init in %s...", cwd))
       output, err := runTofuCmd([]string{"init", "-input=false"}, extraArgs)
       if err != nil {
	       printError(fmt.Sprintf("OpenTofu init failed in %s", cwd), err, output)
	       os.Exit(1)
       }

       printStatus(outputs.Running, fmt.Sprintf("Running tofu validate in %s...", cwd))
       output, err = runTofuCmd([]string{"validate"}, extraArgs)
       if err != nil {
	       printError(fmt.Sprintf("OpenTofu validate failed in %s", cwd), err, output)
	       os.Exit(1)
       }

       printStatus(outputs.ThumbsUp, "OpenTofu validate completed successfully.")
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

// printError prints a colored error message with output
func printError(prefix string, err error, output []byte) {
	fmt.Printf(outputs.EmojiColorText(outputs.Error, "%s: %v\n%s", outputs.Red), prefix, err, output)
}
