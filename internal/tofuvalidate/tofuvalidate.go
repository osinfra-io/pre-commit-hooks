package tofuvalidate

import (
	"os/exec"

	"pre-commit-hooks/internal/testutil"
)

// CheckOpenTofuInstalled delegates to shared testutil implementation.
var CheckOpenTofuInstalled = testutil.CheckOpenTofuInstalled

// RunTofuValidate runs tofu validate in the given directory with extra args.
// Returns output and error.
func RunTofuValidate(dir string, extraArgs []string) (string, error) {
	args := append([]string{"validate"}, extraArgs...)
	cmd := exec.Command("tofu", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	return string(output), err
}
