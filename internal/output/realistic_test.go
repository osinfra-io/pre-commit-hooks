package output

import (
	"fmt"
	"testing"
)

// TestRealisticWarningScenarios demonstrates the warning handling with realistic examples
func TestRealisticWarningScenarios(t *testing.T) {
	// Example 1: Pure warning (like the one from the issue)
	warningOutput := `Warning: Dependency lock file entries automatically updated

OpenTofu automatically rewrote some entries in your dependency lock file:
  - registry.terraform.io/hashicorp/google => registry.opentofu.org/hashicorp/google
  - registry.terraform.io/hashicorp/google-beta => registry.opentofu.org/hashicorp/google-beta
  - registry.terraform.io/hashicorp/random => registry.opentofu.org/hashicorp/random
  - registry.terraform.io/hashicorp/tls => registry.opentofu.org/hashicorp/tls

The version selections were preserved, but the hashes were not because the
OpenTofu project's provider releases are not byte-for-byte identical.`

	result1 := ParseTofuOutput(warningOutput, fmt.Errorf("warning exit"))
	if result1.HasError {
		t.Error("Pure warning should not be treated as error")
	}
	if len(result1.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result1.Warnings))
	}
	if len(result1.Errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(result1.Errors))
	}
	
	// Example 2: Mixed warning and error
	mixedOutput := `Warning: Deprecated feature

This feature is deprecated.

Error: Invalid configuration

  on main.tf line 5:
  5: invalid syntax

Configuration syntax is invalid.`

	result2 := ParseTofuOutput(mixedOutput, fmt.Errorf("error exit"))
	if !result2.HasError {
		t.Error("Mixed warning and error should be treated as error")
	}
	if len(result2.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result2.Warnings))
	}
	if len(result2.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result2.Errors))
	}
	
	// Example 3: Formatting needed (should not be treated as error)
	formatOutput := `File1.tf needs formatting
File2.tf needs formatting`

	result3 := ParseTofuOutput(formatOutput, fmt.Errorf("format exit"))
	if result3.HasError {
		t.Error("Formatting needed should not be treated as error")
	}
	if len(result3.Warnings) != 0 {
		t.Errorf("Expected 0 warnings, got %d", len(result3.Warnings))
	}
	if len(result3.Errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(result3.Errors))
	}
}