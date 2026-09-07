package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"pre-commit-hooks/internal/output"
	"pre-commit-hooks/internal/testutil"
)

func main() {
	err := RunTofuTestCLI(
		os.Args[1:],
		testutil.CheckOpenTofuInstalled,
		os.Getwd,
		hasTestFiles,
		runTofuTest,
		printStatus,
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
) error {
	if !checkInstalled() {
		fmt.Println("OpenTofu is not installed or not in PATH.")
		return fmt.Errorf("OpenTofu not installed")
	}

	rootDir, err := getwd()
	if err != nil {
		fmt.Println("Could not get working directory.")
		return err
	}

	hasTests, err := hasTestFiles(rootDir)
	if err != nil {
		fmt.Printf("Error checking for test files: %v\n", err)
		return err
	}

	if !hasTests {
		printStatus(output.Running, "No OpenTofu test files (.tftest.hcl) found, skipping tests.")
		return nil
	}

	testOutput, err := runTest(rootDir, extraArgs)

	if err != nil {
		fmt.Println()
		c := output.NewCard(output.Red)
		c.Open(output.Badge("FAIL", output.BoldRed), output.Title("OpenTofu test"))
		c.Blank()
		for _, line := range strings.Split(testOutput, "\n") {
			if strings.TrimSpace(line) != "" {
				c.Line(fmt.Sprintf("%s%s%s", output.Dim, line, output.Reset))
			}
		}
		c.Close()
		fmt.Println()
		printStatus(output.Error, "OpenTofu test failed.")
		fmt.Println()
		return fmt.Errorf("test failed: %w", err)
	}

	c := output.NewCard(output.Green)
	c.Open(output.Badge("PASS", output.BoldGreen), output.Title("OpenTofu test"))
	c.Blank()
	for _, line := range strings.Split(testOutput, "\n") {
		if strings.TrimSpace(line) != "" {
			c.Line(fmt.Sprintf("%s%s%s", output.Dim, line, output.Reset))
		}
	}
	c.Close()
	fmt.Println()
	return nil
}

// printStatus prints a colored emoji status message
func printStatus(emoji, msg string) {
	fmt.Println(output.EmojiColorText(emoji, msg, output.Green))
}

func hasTestFiles(rootDir string) (bool, error) {
	found := false
	cleanRoot := filepath.Clean(rootDir)
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := info.Name()
			if filepath.Clean(path) != cleanRoot && strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasSuffix(info.Name(), ".tftest.hcl") {
			found = true
			return filepath.SkipAll
		}

		return nil
	})

	return found, err
}

func runTofuTest(dir string, extraArgs []string) (string, error) {
	args := append([]string{"test"}, extraArgs...)
	cmd := exec.Command("tofu", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	return string(output), err
}
