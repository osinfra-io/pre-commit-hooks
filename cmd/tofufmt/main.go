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

// truncateLine truncates s to maxWidth runes, appending "…" if truncated.
// Lines at or below maxWidth (or when maxWidth <= 1) are returned unchanged.
func truncateLine(s string, maxWidth int) string {
	runes := []rune(s)
	if maxWidth <= 1 || len(runes) <= maxWidth {
		return s
	}
	return string(runes[:maxWidth-1]) + "…"
}

func main() {
	err := RunTofuFmtCLI(
		os.Args[1:],
		testutil.CheckOpenTofuInstalled,
		os.Getwd,
		runTofuFmt,
		formatFiles,
	)
	if err != nil {
		os.Exit(1)
	}
}

func runTofuFmt(dir string, extraArgs []string) (string, error) {
	args := append([]string{"fmt", "-check", "-recursive", "--diff"}, extraArgs...)
	cmd := exec.Command("tofu", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func formatFiles(dir string, extraArgs []string) error {
	args := append([]string{"fmt", "-recursive"}, extraArgs...)
	cmd := exec.Command("tofu", args...)
	cmd.Dir = dir
	return cmd.Run()
}

// RunTofuFmtCLI runs the tofu fmt CLI logic. Returns error if any step fails.
func RunTofuFmtCLI(
	extraArgs []string,
	checkInstalled func() bool,
	getwd func() (string, error),
	runTofuFmt func(string, []string) (string, error),
	formatFiles func(string, []string) error,
) error {
	if !checkInstalled() {
		fmt.Println("OpenTofu is not installed or not in PATH.")
		return fmt.Errorf("OpenTofu not installed")
	}
	wd, err := getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return err
	}
	baseDir := filepath.Base(wd)
	outputStr, err := runTofuFmt(wd, extraArgs)
	fmt.Println()
	if err != nil {
		c := output.NewCard(output.Yellow)
		c.Open(output.Badge("WARNING", output.BoldYellow), output.Title("Unformatted OpenTofu files"))
		c.Line(fmt.Sprintf("%s %s", output.File, output.Colorize(baseDir, output.Gray)))
		c.Blank()
		// "│  " prefix is 3 visible chars; subtract to get usable content width.
		contentWidth := output.TermWidth() - 3
		for _, line := range strings.Split(outputStr, "\n") {
			if strings.TrimSpace(line) != "" {
				var color string
				switch {
				case strings.HasPrefix(line, "+"):
					color = output.Green
				case strings.HasPrefix(line, "-"):
					color = output.Yellow
				default:
					color = output.Dim
				}
				c.Line(fmt.Sprintf("%s%s%s", color, truncateLine(line, contentWidth), output.Reset))
			}
		}
		c.Close()
		fmt.Println()
		fmtErr := formatFiles(wd, extraArgs)
		fmt.Println()
		if fmtErr != nil {
			ec := output.NewCard(output.Red)
			ec.Open(output.Badge("ERROR", output.BoldRed), output.Title("tofu fmt failed"))
			ec.Line(fmt.Sprintf("%s %s", output.File, output.Colorize(baseDir, output.Gray)))
			ec.Blank()
			ec.Line(fmt.Sprintf("%s%s%s", output.Dim, fmtErr.Error(), output.Reset))
			ec.Close()
			return fmtErr
		}
	}
	return nil
}
