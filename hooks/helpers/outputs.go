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
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
)

// Colorize function to wrap text with given color
func Colorize(text, color string) string {
	return fmt.Sprintf("%s%s%s", color, text, Reset)
}

// Helper functions for specific colors
func RedText(text string) string {
	return Colorize(text, Red)
}

func GreenText(text string) string {
	return Colorize(text, Green)
}

func YellowText(text string) string {
	return Colorize(text, Yellow)
}

func BlueText(text string) string {
	return Colorize(text, Blue)
}

func PurpleText(text string) string {
	return Colorize(text, Purple)
}

func CyanText(text string) string {
	return Colorize(text, Cyan)
}

func WhiteText(text string) string {
	return Colorize(text, White)
}

// Emoji helpers
const (
	Error      = "❌"
	Warning    = "🚧"
	Working    = "🔨"
	Running    = "🏃"
	ThumbsUp   = "👍"
	ThumbsDown = "👎"
	Diamond    = "🔸"
)

// Helper functions for text with emojis
func GreenThumbsUp(text string) string {
	return fmt.Sprintf("%s %s", ThumbsUp, GreenText(text))
}

func RedError(text string) string {
	return fmt.Sprintf("%s %s", ThumbsDown, RedText(text))
}

func YellowDiamond(text string) string {
	return fmt.Sprintf("%s %s", Diamond, YellowText(text))
}

func YellowWarning(text string) string {
	return fmt.Sprintf("%s %s", Warning, YellowText(text))
}

func CyanWorking(text string) string {
	return fmt.Sprintf("%s %s", Working, CyanText(text))
}

func CyanRunning(text string) string {
	return fmt.Sprintf("%s %s", Running, CyanText(text))
}
