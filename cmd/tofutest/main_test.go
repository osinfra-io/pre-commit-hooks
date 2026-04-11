package main

import (
	"errors"
	"testing"
)

func TestRunTofuTestCLI_TofuNotInstalled(t *testing.T) {
	called := false
	exitCode := 0
	
	checkInstalled := func() bool { return false }
	getwd := func() (string, error) { return "/fake", nil }
	hasTestFiles := func(string) (bool, error) { return true, nil }
	runTest := func(string, []string) (string, error) { return "", nil }
	printStatus := func(string, string) {}
	exit := func(code int) { 
		exitCode = code
		called = true
	}

	err := RunTofuTestCLI(nil, checkInstalled, getwd, hasTestFiles, runTest, printStatus, exit)
	
	if err == nil {
		t.Error("Expected error when tofu not installed, got nil")
	}
	if !called {
		t.Error("Expected exit to be called")
	}
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}
}

func TestRunTofuTestCLI_GetwdError(t *testing.T) {
	called := false
	exitCode := 0
	
	checkInstalled := func() bool { return true }
	getwd := func() (string, error) { return "", errors.New("getwd failed") }
	hasTestFiles := func(string) (bool, error) { return true, nil }
	runTest := func(string, []string) (string, error) { return "", nil }
	printStatus := func(string, string) {}
	exit := func(code int) { 
		exitCode = code
		called = true
	}

	err := RunTofuTestCLI(nil, checkInstalled, getwd, hasTestFiles, runTest, printStatus, exit)
	
	if err == nil {
		t.Error("Expected error when getwd fails, got nil")
	}
	if !called {
		t.Error("Expected exit to be called")
	}
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}
}

func TestRunTofuTestCLI_NoTestFiles(t *testing.T) {
	called := false
	exitCode := -1
	
	checkInstalled := func() bool { return true }
	getwd := func() (string, error) { return "/fake", nil }
	hasTestFiles := func(string) (bool, error) { return false, nil }
	runTest := func(string, []string) (string, error) { return "", nil }
	printStatus := func(string, string) {}
	exit := func(code int) { 
		exitCode = code
		called = true
	}

	err := RunTofuTestCLI(nil, checkInstalled, getwd, hasTestFiles, runTest, printStatus, exit)
	
	if err != nil {
		t.Errorf("Expected no error when no test files, got %v", err)
	}
	if !called {
		t.Error("Expected exit to be called")
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestRunTofuTestCLI_HasTestFilesError(t *testing.T) {
	called := false
	exitCode := 0
	
	checkInstalled := func() bool { return true }
	getwd := func() (string, error) { return "/fake", nil }
	hasTestFiles := func(string) (bool, error) { return false, errors.New("walk error") }
	runTest := func(string, []string) (string, error) { return "", nil }
	printStatus := func(string, string) {}
	exit := func(code int) { 
		exitCode = code
		called = true
	}

	err := RunTofuTestCLI(nil, checkInstalled, getwd, hasTestFiles, runTest, printStatus, exit)
	
	if err == nil {
		t.Error("Expected error when hasTestFiles fails, got nil")
	}
	if !called {
		t.Error("Expected exit to be called")
	}
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}
}

func TestRunTofuTestCLI_TestSuccess(t *testing.T) {
	checkInstalled := func() bool { return true }
	getwd := func() (string, error) { return "/fake", nil }
	hasTestFiles := func(string) (bool, error) { return true, nil }
	runTest := func(string, []string) (string, error) { return "All tests passed", nil }
	printStatus := func(string, string) {}
	exit := func(code int) { 
		t.Error("Exit should not be called on success")
	}

	err := RunTofuTestCLI(nil, checkInstalled, getwd, hasTestFiles, runTest, printStatus, exit)
	
	if err != nil {
		t.Errorf("Expected no error when tests pass, got %v", err)
	}
}

func TestRunTofuTestCLI_TestFailure(t *testing.T) {
	called := false
	exitCode := 0
	
	checkInstalled := func() bool { return true }
	getwd := func() (string, error) { return "/fake", nil }
	hasTestFiles := func(string) (bool, error) { return true, nil }
	runTest := func(string, []string) (string, error) { return "Test failed", errors.New("test error") }
	printStatus := func(string, string) {}
	exit := func(code int) { 
		exitCode = code
		called = true
	}

	err := RunTofuTestCLI(nil, checkInstalled, getwd, hasTestFiles, runTest, printStatus, exit)
	
	if err == nil {
		t.Error("Expected error when tests fail, got nil")
	}
	if !called {
		t.Error("Expected exit to be called")
	}
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}
}

func TestRunTofuTestCLI_ExtraArgs(t *testing.T) {
	var receivedArgs []string
	
	checkInstalled := func() bool { return true }
	getwd := func() (string, error) { return "/fake", nil }
	hasTestFiles := func(string) (bool, error) { return true, nil }
	runTest := func(dir string, args []string) (string, error) { 
		receivedArgs = args
		return "Tests passed", nil 
	}
	printStatus := func(string, string) {}
	exit := func(code int) {}

	extraArgs := []string{"-verbose", "-filter=TestFoo"}
	err := RunTofuTestCLI(extraArgs, checkInstalled, getwd, hasTestFiles, runTest, printStatus, exit)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(receivedArgs) != 2 {
		t.Errorf("Expected 2 args, got %d", len(receivedArgs))
	}
	if len(receivedArgs) > 0 && receivedArgs[0] != "-verbose" {
		t.Errorf("Expected first arg to be -verbose, got %s", receivedArgs[0])
	}
	if len(receivedArgs) > 1 && receivedArgs[1] != "-filter=TestFoo" {
		t.Errorf("Expected second arg to be -filter=TestFoo, got %s", receivedArgs[1])
	}
}

func TestParseExtraArgs_StandardFlag(t *testing.T) {
	got := parseExtraArgs([]string{"-verbose"})
	if len(got) != 1 || got[0] != "-verbose" {
		t.Errorf("parseExtraArgs([-verbose]) = %v, want [-verbose]", got)
	}
}

func TestParseExtraArgs_EqualForm(t *testing.T) {
	got := parseExtraArgs([]string{"-filter=TestFoo"})
	if len(got) != 1 || got[0] != "-filter=TestFoo" {
		t.Errorf("parseExtraArgs([-filter=TestFoo]) = %v, want [-filter=TestFoo]", got)
	}
}

func TestParseExtraArgs_SplitForm(t *testing.T) {
	got := parseExtraArgs([]string{"-filter", "TestFoo"})
	if len(got) != 2 || got[0] != "-filter" || got[1] != "TestFoo" {
		t.Errorf("parseExtraArgs([-filter TestFoo]) = %v, want [-filter TestFoo]", got)
	}
}

func TestParseExtraArgs_Mixed(t *testing.T) {
	got := parseExtraArgs([]string{"-verbose", "-filter", "TestFoo", "-json"})
	want := []string{"-verbose", "-filter", "TestFoo", "-json"}
	if len(got) != len(want) {
		t.Fatalf("parseExtraArgs() = %v, want %v", got, want)
	}
	for i, v := range want {
		if got[i] != v {
			t.Errorf("parseExtraArgs()[%d] = %q, want %q", i, got[i], v)
		}
	}
}

func TestParseExtraArgs_NonFlagTokensOnly(t *testing.T) {
	got := parseExtraArgs([]string{"file1.tf", "file2.tofu"})
	if len(got) != 0 {
		t.Errorf("parseExtraArgs() = %v, want empty slice", got)
	}
}

func TestParseExtraArgs_TrailingFlagNoValue(t *testing.T) {
	got := parseExtraArgs([]string{"-verbose", "-filter"})
	want := []string{"-verbose", "-filter"}
	if len(got) != len(want) {
		t.Fatalf("parseExtraArgs() = %v, want %v", got, want)
	}
	for i, v := range want {
		if got[i] != v {
			t.Errorf("parseExtraArgs()[%d] = %q, want %q", i, got[i], v)
		}
	}
}

func TestParseExtraArgs_BoolFlagDoesNotCapturePositional(t *testing.T) {
	// Boolean flags must not consume a following positional token.
	got := parseExtraArgs([]string{"-verbose", "file.tf"})
	if len(got) != 1 || got[0] != "-verbose" {
		t.Errorf("parseExtraArgs([-verbose file.tf]) = %v, want [-verbose]", got)
	}
}

func TestParseExtraArgs_EndOfFlags(t *testing.T) {
	// A "--" token must end flag processing; everything after is ignored.
	got := parseExtraArgs([]string{"-verbose", "--", "-filter", "TestFoo"})
	if len(got) != 1 || got[0] != "-verbose" {
		t.Errorf("parseExtraArgs([-verbose -- -filter TestFoo]) = %v, want [-verbose]", got)
	}
}
