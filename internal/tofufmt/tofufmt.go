package tofufmt

import (
	"os/exec"

	"pre-commit-hooks/internal/testutil"
)

// CheckOpenTofuInstalled delegates to shared testutil implementation.
var CheckOpenTofuInstalled = testutil.CheckOpenTofuInstalled

// RunTofuFmt runs tofu fmt recursively in the given directory with extra args.
// Returns output and error.
func RunTofuFmt(dir string, extraArgs []string) (string, error) {
	args := append([]string{"fmt", "-check", "-recursive", "--diff"}, extraArgs...)
	cmd := exec.Command("tofu", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// FormatFiles runs tofu fmt to format files in the given directory with extra args.
func FormatFiles(dir string, extraArgs []string) error {
	args := append([]string{"fmt", "-recursive"}, extraArgs...)
	cmd := exec.Command("tofu", args...)
	cmd.Dir = dir
	return cmd.Run()
}
