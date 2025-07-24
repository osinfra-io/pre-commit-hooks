package output

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestPrintWarningSummary(t *testing.T) {
	msgs := []TofuMessage{
		{Step: "init", RelPath: "dir1", Output: "Warning: something happened\nDetails here"},
		{Step: "validate", RelPath: "dir2", Output: "Warning: another warning\nMore details"},
	}
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	PrintWarningSummary(msgs)
	w.Close()
	outBytes, _ := io.ReadAll(r)
	output := string(outBytes)
	if !strings.Contains(output, "Warning Summary:") {
		t.Errorf("Expected warning summary header, got: %s", output)
	}
	if !strings.Contains(output, "OpenTofu init warning in: dir1") || !strings.Contains(output, "OpenTofu validate warning in: dir2") {
		t.Errorf("Expected warning details for both messages, got: %s", output)
	}
}

func TestPrintErrorSummary(t *testing.T) {
	msgs := []TofuMessage{
		{Step: "validate", RelPath: "dir3", Output: "Error: failed validation\nDetails"},
	}
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	PrintErrorSummary(msgs, func(out string, _ bool) { os.Stdout.Write([]byte(out)) })
	w.Close()
	outBytes, _ := io.ReadAll(r)
	output := string(outBytes)
	if !strings.Contains(output, "Error Summary:") {
		t.Errorf("Expected error summary header, got: %s", output)
	}
	if !strings.Contains(output, "OpenTofu validate failed in: dir3") {
		t.Errorf("Expected error details, got: %s", output)
	}
}
