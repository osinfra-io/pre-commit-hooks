package output

import (
	"fmt"
	"regexp"
	"strings"
)

// ANSI escape codes for color
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
)

// Colorize function to wrap text with given color
func Colorize(text, color string) string {
	return fmt.Sprintf("%s%s%s", color, text, Reset)
}

// Emoji constants
const (
	// Error: Indicates a failure or error condition
	Error = "💀"
	// Warning: Indicates a warning or something that needs attention
	Warning = "🚧"
	// Running: Indicates an ongoing process or operation
	Running = "⚙️"
	// ThumbsUp: Indicates success or positive outcome
	ThumbsUp = "👍"
)

// Generic function to combine emoji and colored text
func EmojiColorText(emoji, text, color string) string {
	return fmt.Sprintf("%s %s", emoji, Colorize(text, color))
}

// Example usage:
// fmt.Println(EmojiColorText(ThumbsUp, "All OpenTofu files are formatted.", Green))
// fmt.Println(EmojiColorText(Warning, "Found unformatted OpenTofu files:", Yellow))

// ParsedOutput represents the result of parsing tofu command output
type ParsedOutput struct {
	Warnings []string // Warning messages found in the output
	Errors   []string // Error messages found in the output
	HasError bool     // True if there are actual errors (not just warnings)
}

// ParseTofuOutput parses tofu command output to distinguish warnings from errors
func ParseTofuOutput(output string, cmdErr error) ParsedOutput {
	parsed := ParsedOutput{
		Warnings: []string{},
		Errors:   []string{},
		HasError: false,
	}

	if output == "" && cmdErr == nil {
		return parsed
	}

	// Split output into lines for processing
	lines := strings.Split(strings.TrimSpace(output), "\n")
	
	// Regex patterns to identify warnings and errors
	warningPattern := regexp.MustCompile(`(?i)^\s*warning:`)
	errorPattern := regexp.MustCompile(`(?i)^\s*error:`)
	
	var currentWarning []string
	var currentError []string
	inWarning := false
	inError := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			// Empty lines continue the current context but don't get added
			continue
		}
		
		// Check if this line starts a warning
		if warningPattern.MatchString(line) {
			// Save previous warning if any
			if inWarning && len(currentWarning) > 0 {
				parsed.Warnings = append(parsed.Warnings, strings.Join(currentWarning, "\n"))
			}
			// Save previous error if any (when switching from error to warning)
			if inError && len(currentError) > 0 {
				parsed.Errors = append(parsed.Errors, strings.Join(currentError, "\n"))
				currentError = []string{}
			}
			// Start new warning
			currentWarning = []string{line}
			inWarning = true
			inError = false
			continue
		}
		
		// Check if this line starts an error
		if errorPattern.MatchString(line) {
			// Save previous warning if any (when switching from warning to error)
			if inWarning && len(currentWarning) > 0 {
				parsed.Warnings = append(parsed.Warnings, strings.Join(currentWarning, "\n"))
				currentWarning = []string{}
			}
			// Save previous error if any
			if inError && len(currentError) > 0 {
				parsed.Errors = append(parsed.Errors, strings.Join(currentError, "\n"))
			}
			// Start new error
			currentError = []string{line}
			inError = true
			inWarning = false
			parsed.HasError = true
			continue
		}
		
		// Add line to current context (warning or error)
		if inWarning {
			currentWarning = append(currentWarning, line)
		} else if inError {
			currentError = append(currentError, line)
		}
		// Note: We removed the automatic error assignment here
	}
	
	// Save final warning or error
	if inWarning && len(currentWarning) > 0 {
		parsed.Warnings = append(parsed.Warnings, strings.Join(currentWarning, "\n"))
	}
	if inError && len(currentError) > 0 {
		parsed.Errors = append(parsed.Errors, strings.Join(currentError, "\n"))
	}
	
	// Only treat command error as actual error if we didn't find any warnings
	// and the output suggests it's a real error (not just formatting needed)
	if cmdErr != nil && !parsed.HasError && len(parsed.Warnings) == 0 {
		// For certain types of output, we can infer it's a real error
		lowerOutput := strings.ToLower(output)
		if strings.Contains(lowerOutput, "failed") || 
		   strings.Contains(lowerOutput, "invalid") ||
		   strings.Contains(lowerOutput, "syntax error") ||
		   (output != "" && !strings.Contains(lowerOutput, "format")) {
			parsed.Errors = append(parsed.Errors, output)
			parsed.HasError = true
		}
	}
	
	return parsed
}

// PrintWarnings prints a formatted list of warnings
func PrintWarnings(warnings []string) {
	if len(warnings) == 0 {
		return
	}
	
	fmt.Println(EmojiColorText(Warning, "Warnings found:", Yellow))
	fmt.Println()
	for i, warning := range warnings {
		fmt.Printf("  Warning %d:\n", i+1)
		// Indent each line of the warning
		warningLines := strings.Split(warning, "\n")
		for _, line := range warningLines {
			if strings.TrimSpace(line) != "" {
				fmt.Printf("    %s\n", line)
			}
		}
		fmt.Println()
	}
}
