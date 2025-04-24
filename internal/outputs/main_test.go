package outputs

import (
    "testing"
)


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
