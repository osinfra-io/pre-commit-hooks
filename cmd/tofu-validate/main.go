package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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
	relPath := cwd
	gitRootCmd := exec.Command("git", "rev-parse", "--show-toplevel")
	gitRootOut, gitRootErr := gitRootCmd.Output()
	if gitRootErr == nil {
		gitRoot := string(gitRootOut)
		gitRoot = gitRoot[:len(gitRoot)-1] // remove trailing newline
		// Use filepath.Rel for robust relative path calculation
		if rel, relErr := relPathFromGitRoot(gitRoot, cwd); relErr == nil {
			relPath = rel
		}
	}

	fmt.Println(outputs.EmojiColorText(outputs.Running, fmt.Sprintf("Running tofu init in %s...", relPath), outputs.Green))
	// Collect extra args from .pre-commit-config.yaml via os.Args[1:]
	extraArgs := os.Args[1:]
	initArgs := []string{"init", "-input=false"}
	// Append extra args (e.g., "-upgrade") if provided
	initArgs = append(initArgs, extraArgs...)
	initCmd := exec.Command("tofu", initArgs...)
	_, err = initCmd.CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			fmt.Printf(outputs.EmojiColorText(outputs.Error, "OpenTofu init failed: %v\n", outputs.Red), err)
		} else {
			fmt.Printf(outputs.EmojiColorText(outputs.Error, "Error running tofu init: %v\n", outputs.Red), err)
		}
		os.Exit(1)
	}

	fmt.Println(outputs.EmojiColorText(outputs.Running, fmt.Sprintf("Running tofu validate in %s...", relPath), outputs.Green))
	// Run tofu validate with extra args
	validateArgs := []string{"validate"}
	validateArgs = append(validateArgs, extraArgs...)
	cmd := exec.Command("tofu", validateArgs...)
	_, err = cmd.CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			fmt.Printf(outputs.EmojiColorText(outputs.Error, "OpenTofu validate failed: %v\n", outputs.Red), err)
		} else {
			fmt.Printf(outputs.EmojiColorText(outputs.Error, "Error running tofu validate: %v\n", outputs.Red), err)
		}
		os.Exit(1)
	}

	fmt.Println(outputs.EmojiColorText(outputs.ThumbsUp, "OpenTofu validate completed successfully.", outputs.Green))
}

// relPathFromGitRoot returns the relative path from gitRoot to cwd, or "." if the same
func relPathFromGitRoot(gitRoot, cwd string) (string, error) {
	rel, err := filepath.Rel(gitRoot, cwd)
	if err != nil || rel == "" {
		return filepath.Base(cwd), nil
	}
	// If rel is ".", return the base directory name
	if rel == "." {
		return filepath.Base(cwd), nil
	}
	// Otherwise, return the last element of the relative path
	return filepath.Base(rel), nil
}
