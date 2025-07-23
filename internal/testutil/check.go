package testutil

import "os/exec"

// CheckOpenTofuInstalled returns true if the 'tofu' binary is found in PATH.
func CheckOpenTofuInstalled() bool {
	_, err := exec.LookPath("tofu")
	return err == nil
}
