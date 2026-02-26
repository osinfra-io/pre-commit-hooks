package tofutest

import (
"os"
"os/exec"
"path/filepath"
"strings"

"pre-commit-hooks/internal/testutil"
)

// CheckOpenTofuInstalled delegates to shared testutil implementation.
var CheckOpenTofuInstalled = testutil.CheckOpenTofuInstalled

// HasTestFiles recursively searches for .tftest.hcl files in the given directory.
// Returns true if any test files are found, false otherwise.
func HasTestFiles(rootDir string) (bool, error) {
found := false
err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
if err != nil {
return err
}

// Skip hidden directories and .terraform
if info.IsDir() {
name := info.Name()
if strings.HasPrefix(name, ".") || name == ".terraform" {
return filepath.SkipDir
}
return nil
}

// Check if file has .tftest.hcl extension
if strings.HasSuffix(info.Name(), ".tftest.hcl") {
found = true
return filepath.SkipAll // Stop searching once we find one
}

return nil
})

return found, err
}

// RunTofuTest runs tofu test in the given directory with extra args.
// Returns output and error.
func RunTofuTest(dir string, extraArgs []string) (string, error) {
args := append([]string{"test"}, extraArgs...)
cmd := exec.Command("tofu", args...)
cmd.Dir = dir
output, err := cmd.CombinedOutput()
return string(output), err
}
