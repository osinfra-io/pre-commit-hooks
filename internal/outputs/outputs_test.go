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
