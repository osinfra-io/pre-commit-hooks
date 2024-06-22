package main

import (
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"
)

func TestFormatTerraformFiles(t *testing.T) {
    testDir := filepath.Join("..", "..", "tests", "terraform", "format", "unformatted")
    err := os.Chdir(testDir)
    if err != nil {
        t.Fatalf("Failed to change directory: %v", err)
    }

    cmd := exec.Command("../../../../hooks/terraform/format")
    output, err := cmd.CombinedOutput()

    // If no error, check for the success message
    if err == nil {
        successMessage := "All Terraform files are formatted."
        if !strings.Contains(string(output), successMessage) {
            t.Errorf("Expected output to contain %q, indicating success, but got: %q", successMessage, string(output))
        }
    } else {
        // Handle the case where there is an error (e.g., unformatted files)
        expectedOutput := "Please format the files listed above using 'terraform fmt' before committing."
        if !strings.Contains(string(output), expectedOutput) {
            t.Errorf("Expected output to contain %q, got %q", expectedOutput, string(output))
        }
    }
}
