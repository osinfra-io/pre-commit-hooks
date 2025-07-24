package output

import (
	"fmt"
	"strings"
	"testing"
)

func TestEmojiColorText(t *testing.T) {
	cases := []struct {
		name  string
		emoji string
		text  string
		color string
		want  string
	}{
		{"Red with Error", Error, "fail", Red, Error + " " + Colorize("fail", Red)},
		{"Green with ThumbsUp", ThumbsUp, "ok", Green, ThumbsUp + " " + Colorize("ok", Green)},
		{"Yellow with Warning", Warning, "warn", Yellow, Warning + " " + Colorize("warn", Yellow)},
	}
	for _, c := range cases {
		got := EmojiColorText(c.emoji, c.text, c.color)
		if got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

func TestResetColor(t *testing.T) {
	// Test Reset constant value
	if Reset != "\033[0m" {
		t.Errorf("Reset constant has incorrect value. got: %q, want: %q",
			Reset, "\033[0m")
	}

	// Test Reset color usage
	testText := "test text"
	result := Colorize(testText, Reset)
	expected := "\033[0m" + testText + "\033[0m"

	if result != expected {
		t.Errorf("Reset color formatting incorrect.\ngot: %q\nwant: %q",
			result, expected)
	}
}

func TestColorize_AllColors(t *testing.T) {
	cases := []struct {
		name  string
		color string
		want  string
	}{
		{"Red", Red, "\033[31mtest\033[0m"},
		{"Green", Green, "\033[32mtest\033[0m"},
		{"Yellow", Yellow, "\033[33mtest\033[0m"},
	}
	for _, c := range cases {
		got := Colorize("test", c.color)
		if got != c.want {
			t.Errorf("Colorize(%s): got %q, want %q", c.name, got, c.want)
		}
	}
}

func TestParseTofuOutput(t *testing.T) {
	cases := []struct {
		name           string
		output         string
		cmdErr         error
		expectWarnings int
		expectErrors   int
		expectHasError bool
	}{
		{
			name:           "no output, no error",
			output:         "",
			cmdErr:         nil,
			expectWarnings: 0,
			expectErrors:   0,
			expectHasError: false,
		},
		{
			name: "warning only",
			output: `Warning: Dependency lock file entries automatically updated

OpenTofu automatically rewrote some entries in your dependency lock file:
  - registry.terraform.io/hashicorp/google => registry.opentofu.org/hashicorp/google

The version selections were preserved, but the hashes were not because the
OpenTofu project's provider releases are not byte-for-byte identical.`,
			cmdErr:         nil,
			expectWarnings: 1,
			expectErrors:   0,
			expectHasError: false,
		},
		{
			name: "error only",
			output: `Error: Invalid resource type

  on main.tf line 10:
  10: resource "invalid_resource" "test" {

The provider does not support resource type "invalid_resource".`,
			cmdErr:         fmt.Errorf("command failed"),
			expectWarnings: 0,
			expectErrors:   1,
			expectHasError: true,
		},
		{
			name: "warning and error",
			output: `Warning: Deprecated feature

This feature is deprecated and will be removed in a future version.

Error: Configuration error

  on main.tf line 5:
  5: invalid syntax here

Invalid configuration syntax.`,
			cmdErr:         fmt.Errorf("command failed"),
			expectWarnings: 1,
			expectErrors:   1,
			expectHasError: true,
		},
		{
			name: "multiple warnings",
			output: `Warning: First warning

First warning description.

Warning: Second warning

Second warning description.`,
			cmdErr:         nil,
			expectWarnings: 2,
			expectErrors:   0,
			expectHasError: false,
		},
		{
			name: "command error without explicit error message",
			output: `Some generic failure output
that doesn't have explicit Error: prefix`,
			cmdErr:         fmt.Errorf("command failed"),
			expectWarnings: 0,
			expectErrors:   1,
			expectHasError: true,
		},
		{
			name: "command error for formatting (should not be treated as error)",
			output: `needs format`,
			cmdErr:         fmt.Errorf("command failed"),
			expectWarnings: 0,
			expectErrors:   0,
			expectHasError: false,
		},
		{
			name:           "command error with empty output",
			output:         "",
			cmdErr:         fmt.Errorf("command failed"),
			expectWarnings: 0,
			expectErrors:   0,
			expectHasError: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := ParseTofuOutput(tc.output, tc.cmdErr)

			if len(result.Warnings) != tc.expectWarnings {
				t.Errorf("Expected %d warnings, got %d. Warnings: %v", 
					tc.expectWarnings, len(result.Warnings), result.Warnings)
			}

			if len(result.Errors) != tc.expectErrors {
				t.Errorf("Expected %d errors, got %d. Errors: %v", 
					tc.expectErrors, len(result.Errors), result.Errors)
			}

			if result.HasError != tc.expectHasError {
				t.Errorf("Expected HasError=%v, got %v", tc.expectHasError, result.HasError)
			}

			// Validate warning content for the warning-only case
			if tc.name == "warning only" && len(result.Warnings) > 0 {
				warning := result.Warnings[0]
				if !strings.Contains(warning, "Dependency lock file entries") {
					t.Errorf("Warning content doesn't match expected. Got: %s", warning)
				}
			}
		})
	}
}

func TestPrintWarnings(t *testing.T) {
	// Test with no warnings - should not panic or print anything
	PrintWarnings([]string{})

	// Test with warnings - mainly checking it doesn't panic
	warnings := []string{
		"Warning: First warning\nWith multiple lines",
		"Warning: Second warning\nAlso with multiple lines",
	}
	PrintWarnings(warnings) // Should not panic
}
