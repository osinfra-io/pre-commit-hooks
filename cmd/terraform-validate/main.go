package main

import (
    "fmt"
    "os"
    "os/exec"

    "github.com/osinfra-io/pre-commit-hooks/internal/outputs"
)

func checkTerraformInstalled() bool {
    _, err := exec.LookPath("terraform")
    return err == nil
}

func main() {
    if !checkTerraformInstalled() {
        fmt.Println("Terraform is not installed or not in PATH.")
        os.Exit(1)
    }

    fmt.Println(outputs.EmojiColorText(outputs.Running, "Running terraform validate...", outputs.Purple))

    // Run terraform validate
    cmd := exec.Command("terraform", "validate")
    _, err := cmd.CombinedOutput()
    if err != nil {
        // Check if the error is an ExitError
        if _, ok := err.(*exec.ExitError); ok {
            // Handle specific exit codes if necessary
            fmt.Printf(outputs.EmojiColorText(outputs.Error, "Terraform validate failed: %v\n", outputs.Red), err)
        } else {
            fmt.Printf(outputs.EmojiColorText(outputs.Error, "Error running terraform validate: %v\n", outputs.Red), err)
        }
        os.Exit(1)
    }

    fmt.Println(outputs.EmojiColorText(outputs.ThumbsUp, "Terraform validate completed successfully.", outputs.Green))
}
