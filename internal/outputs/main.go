package outputs

import (
	"fmt"
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
	Error      = "💀"
	Warning    = "🚧"
	Working    = "🔨"
	Running    = "🔩"
	ThumbsUp   = "👍"
	ThumbsDown = "👎"
	Diamond    = "🔸"
)

// Generic function to combine emoji and colored text
func EmojiColorText(emoji, text, color string) string {
	return fmt.Sprintf("%s %s", emoji, Colorize(text, color))
}

// Example usage:
// fmt.Println(EmojiColorText(ThumbsUp, "All OpenTofu files are formatted.", Green))
// fmt.Println(EmojiColorText(Warning, "Found unformatted OpenTofu files:", Yellow))
