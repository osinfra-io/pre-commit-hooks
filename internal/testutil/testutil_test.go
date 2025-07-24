package testutil

import (
	"testing"
)

func TestCheckOpenTofuInstalled(t *testing.T) {
	SkipIfTofuNotInstalled(t)
	if !CheckOpenTofuInstalled() {
		t.Error("CheckOpenTofuInstalled should return true when tofu is installed")
	}
}
