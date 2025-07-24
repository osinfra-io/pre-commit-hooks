package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckOpenTofuInstalled(t *testing.T) {
	SkipIfTofuNotInstalled(t)
	if !CheckOpenTofuInstalled() {
		t.Error("CheckOpenTofuInstalled should return true when tofu is installed")
	}
}

func TestCreateTempDir(t *testing.T) {
	dir, cleanup := CreateTempDir(t, "testutil_temp_")
	defer cleanup()
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Temp dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("Expected a directory, got something else: %v", dir)
	}
}

func TestRestoreWorkingDir(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working dir: %v", err)
	}
	dir, cleanup := CreateTempDir(t, "testutil_restore_")
	defer cleanup()
	restore := RestoreWorkingDir(t, dir)
	newDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get new working dir: %v", err)
	}
	newDirEval, err := filepath.EvalSymlinks(newDir)
	if err != nil {
		t.Fatalf("EvalSymlinks failed for newDir: %v", err)
	}
	dirEval, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatalf("EvalSymlinks failed for dir: %v", err)
	}
	if newDirEval != dirEval {
		t.Errorf("Did not change to new dir. got: %v, want: %v", newDirEval, dirEval)
	}
	restore()
	finalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get restored working dir: %v", err)
	}
	finalDirEval, err := filepath.EvalSymlinks(finalDir)
	if err != nil {
		t.Fatalf("EvalSymlinks failed for finalDir: %v", err)
	}
	origEval, err := filepath.EvalSymlinks(orig)
	if err != nil {
		t.Fatalf("EvalSymlinks failed for orig: %v", err)
	}
	if finalDirEval != origEval {
		t.Errorf("Did not restore to original dir. got: %v, want: %v", finalDirEval, origEval)
	}
}

func TestRunCmd(t *testing.T) {
	output, err := RunCmd("echo", "hello world")
	if err != nil {
		t.Fatalf("RunCmd failed: %v", err)
	}
	if want := "hello world\n"; output != want {
		t.Errorf("RunCmd output: got %q, want %q", output, want)
	}
}
