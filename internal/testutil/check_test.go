package testutil

import (
	"os/exec"
	"testing"
)

func TestCheckOpenTofuInstalled(t *testing.T) {
	if _, err := exec.LookPath("tofu"); err == nil {
		if !CheckOpenTofuInstalled() {
			t.Error("CheckOpenTofuInstalled should return true when tofu is installed")
		}
	} else {
		t.Skip("Skipping positive test as tofu is not installed")
	}

	t.Log("Note: Unable to directly test the case where tofu is not installed")
}
