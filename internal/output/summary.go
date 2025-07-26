package output

import (
	"fmt"
	"strings"
)

type TofuMessage struct {
	Step    string
	RelPath string
	Output  string
}

func PrintWarningSummary(warningMessages []TofuMessage) {
	if len(warningMessages) == 0 {
		return
	}
	fmt.Println(EmojiColorText("⚠️", "Warning Summary:", Yellow))
	fmt.Println()
	for _, msg := range warningMessages {
		fmt.Printf(EmojiColorText(Warning, "OpenTofu %s warning in: %s\n", Yellow), msg.Step, msg.RelPath)
		lines := strings.Split(msg.Output, "\n")
		inWarning := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.Contains(trimmed, "Warning:") {
				inWarning = true
			}
			if inWarning {
				fmt.Printf("    %s\n", line)
			}
		}
	}
}

func PrintErrorSummary(errorMessages []TofuMessage, printIndentedOutput func(string, bool)) {
	if len(errorMessages) == 0 {
		return
	}
	fmt.Println(EmojiColorText("❗", "Error Summary:", Red))
	fmt.Println()
	for _, msg := range errorMessages {
		fmt.Printf(EmojiColorText(Error, "OpenTofu %s failed in: %s\n", Red), msg.Step, msg.RelPath)
		printIndentedOutput(msg.Output, false)
	}
}
