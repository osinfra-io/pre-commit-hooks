package main

import (
	"fmt"
	"os"
	"strings"

	"pre-commit-hooks/internal/output"
	tofutest "pre-commit-hooks/internal/tofutest"
)

func main() {
	err := RunTofuTestCLI(
		parseExtraArgs(os.Args[1:]),
		tofutest.CheckOpenTofuInstalled,
		os.Getwd,
		tofutest.HasTestFiles,
		tofutest.RunTofuTest,
		printStatus,
		os.Exit,
	)
	if err != nil {
		os.Exit(1)
	}
}

// RunTofuTestCLI runs the tofu test CLI logic. Returns error if any step fails.
func RunTofuTestCLI(
	extraArgs []string,
	checkInstalled func() bool,
	getwd func() (string, error),
	hasTestFiles func(string) (bool, error),
	runTest func(string, []string) (string, error),
	printStatus func(string, string),
	exit func(int),
) error {
	if !checkInstalled() {
		fmt.Println("OpenTofu is not installed or not in PATH.")
		exit(1)
		return fmt.Errorf("OpenTofu not installed")
	}

	rootDir, err := getwd()
	if err != nil {
		fmt.Println("Could not get working directory.")
		exit(1)
		return err
	}

	hasTests, err := hasTestFiles(rootDir)
	if err != nil {
		fmt.Printf("Error checking for test files: %v\n", err)
		exit(1)
		return err
	}

	if !hasTests {
		printStatus(output.Running, "No OpenTofu test files (.tftest.hcl) found, skipping tests.")
		exit(0)
		return nil
	}

	printStatus(output.Running, "Running tofu test...")
	testOutput, err := runTest(rootDir, extraArgs)

	// Print the output
	printIndentedOutput(testOutput, true)

	if err != nil {
		printStatus(output.Error, "OpenTofu test failed.")
		fmt.Println()
		exit(1)
		return fmt.Errorf("test failed")
	}

	printStatus(output.ThumbsUp, "OpenTofu test completed successfully.")
	fmt.Println()
	return nil
}

// printIndentedOutput prints each line of output indented for better readability
func printIndentedOutput(output string, addNewline bool) {
	lines := strings.Split(output, "\n")
	lastNonEmpty := -1
	for idx := range lines {
		if strings.TrimSpace(lines[idx]) != "" {
			lastNonEmpty = idx
		}
	}
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			fmt.Printf("    %s\n", line)
		}
	}
	// Only add newline if not already present at the end
	if addNewline && lastNonEmpty != len(lines)-1 {
		fmt.Println()
	}
}

// printStatus prints a colored emoji status message
func printStatus(emoji, msg string) {
	fmt.Println(output.EmojiColorText(emoji, msg, output.Green))
}

// parseExtraArgs filters os.Args tokens, keeping only flags (tokens starting with '-')
// and their values. Equals-form flags (-flag=value) are kept as a single token.
// Split-form flags (-flag value) are kept as two tokens.
func parseExtraArgs(args []string) []string {
	extraArgs := []string{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			extraArgs = append(extraArgs, arg)
			// If this flag doesn't contain '=' and the next token exists and doesn't
			// start with '-', it's a split-form flag — include the value too.
			if !strings.Contains(arg, "=") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				extraArgs = append(extraArgs, args[i+1])
				i++ // skip the value token
			}
		}
	}
	return extraArgs
}
