package main

import (
	"fmt"
	"os"
	"strings"

	"pre-commit-hooks/internal/output"
	tofu_fmt "pre-commit-hooks/internal/tofufmt"
)

func main() {
	err := RunTofuFmtCLI(
		os.Args[1:],
		os.Getwd,
		tofu_fmt.RunTofuFmt,
		tofu_fmt.FormatFiles,
	)
	if err != nil {
		os.Exit(1)
	}
}

// RunTofuFmtCLI runs the tofu fmt CLI logic. Returns error if any step fails.
func RunTofuFmtCLI(
	extraArgs []string,
	getwd func() (string, error),
	runTofuFmt func(string, []string) (string, error),
	formatFiles func(string, []string) error,
) error {
	if !tofu_fmt.CheckOpenTofuInstalled() {
		fmt.Println("OpenTofu is not installed or not in PATH.")
		return fmt.Errorf("OpenTofu not installed")
	}
	wd, err := getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return err
	}
	dirName := wd[strings.LastIndex(wd, string(os.PathSeparator))+1:]
	fmt.Printf("Running tofu fmt recursively in directory: %s\n", dirName)

	printStatus(output.Running, "Running tofu fmt...")

	outputStr, err := runTofuFmt(wd, extraArgs)
	fmt.Println()
	if err != nil {
		fmt.Println(output.EmojiColorText(output.Warning, "Found unformatted OpenTofu files:", output.Yellow))
		fmt.Println(outputStr)
		printStatus(output.Running, "Formatting files with tofu fmt...")
		fmtErr := formatFiles(wd, extraArgs)
		fmt.Println()
		if fmtErr != nil {
			fmt.Println(output.EmojiColorText(output.Error, "Error running tofu fmt:", output.Red))
			fmt.Println(fmtErr)
			return fmtErr
		}
		printStatus(output.ThumbsUp, "Files formatted successfully with tofu fmt.")
		fmt.Println()
	} else {
		printStatus(output.ThumbsUp, "All OpenTofu files are formatted.")
		fmt.Println()
	}
	return nil
}

// printStatus prints a colored emoji status message
func printStatus(emoji, msg string) {
	fmt.Println(output.EmojiColorText(emoji, msg, output.Green))
}
