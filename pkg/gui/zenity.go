package gui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

const (
	// DefaultDialogWidth is the default width for dialog windows in pixels
	DefaultDialogWidth = 400
	// MaxLineLength is the maximum length of a line in characters before it's wrapped
	MaxLineLength = 100
)

// DialogProvider defines the interface for displaying user input dialogs
type DialogProvider interface {
	ShowInputDialog(prompt string, title string) (string, error)
}

// ZenityDialog provides dialog functionality using zenity on Linux and osascript on macOS
type ZenityDialog struct {
	// Width in pixels for the dialog window
	Width int
}

// NewZenityDialog creates a new ZenityDialog with default settings
func NewZenityDialog() *ZenityDialog {
	return &ZenityDialog{
		Width: DefaultDialogWidth,
	}
}

// NewZenityDialogWithWidth creates a new ZenityDialog with the specified width
func NewZenityDialogWithWidth(width int) *ZenityDialog {
	return &ZenityDialog{
		Width: width,
	}
}

// ShowInputDialog displays an input dialog with the given prompt and title
// Returns the user's input or an error if the dialog could not be displayed
// or the user cancelled the dialog
func (z *ZenityDialog) ShowInputDialog(prompt string, title string) (string, error) {
	switch runtime.GOOS {
	case "linux":
		return z.linuxDialog(prompt, title)
	case "darwin":
		return z.macOSDialog(prompt, title)
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// linuxDialog shows a dialog using zenity on Linux
func (z *ZenityDialog) linuxDialog(prompt string, title string) (string, error) {
	// Ensure width is reasonable
	width := z.Width
	if width <= 0 {
		width = DefaultDialogWidth
	}

	// Force line wrapping by inserting newlines
	formattedPrompt := formatTextWithLinebreaks(prompt, MaxLineLength)

	// Build command with width parameter to ensure text wrapping
	// Use --entry for single-line input that supports Enter key submission
	cmd := exec.Command(
		"zenity",
		"--entry",
		"--title", title,
		"--text", formattedPrompt,
		"--width", fmt.Sprintf("%d", width),
	)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Exit code 1 typically means user cancelled
			if exitErr.ExitCode() == 1 {
				return "", fmt.Errorf("user cancelled input")
			}
		}
		return "", fmt.Errorf("error showing dialog: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// macOSDialog shows a dialog using osascript on macOS
func (z *ZenityDialog) macOSDialog(prompt string, title string) (string, error) {
	// Format prompt with manual line breaks for better display
	formattedPrompt := formatTextWithLinebreaks(prompt, MaxLineLength)

	// AppleScript to display a dialog with input field
	script := fmt.Sprintf(`osascript -e 'display dialog "%s" with title "%s" default answer "" buttons {"Cancel", "OK"} default button "OK"' -e 'text returned of result'`,
		escapeAppleScriptString(formattedPrompt),
		escapeAppleScriptString(title))

	cmd := exec.Command("bash", "-c", script)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// For macOS, dialog cancellation returns exit code 1
			if exitErr.ExitCode() == 1 {
				return "", fmt.Errorf("user cancelled input")
			}
		}
		return "", fmt.Errorf("error showing dialog: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// formatTextWithLinebreaks inserts line breaks to improve readability
func formatTextWithLinebreaks(text string, maxLineLength int) string {
	if maxLineLength <= 0 {
		maxLineLength = MaxLineLength
	}

	// Split the input text by existing newlines if any
	paragraphs := strings.Split(text, "\n")
	var result strings.Builder

	for i, paragraph := range paragraphs {
		if i > 0 {
			result.WriteString("\n")
		}

		words := strings.Fields(paragraph)
		if len(words) == 0 {
			continue
		}

		lineLength := 0

		for j, word := range words {
			// If adding this word would exceed the line length, add a newline
			if lineLength > 0 && lineLength+len(word)+1 > maxLineLength {
				result.WriteString("\n")
				lineLength = 0
			} else if j > 0 {
				// Add a space before the word if it's not the first word on the line
				result.WriteString(" ")
				lineLength++
			}

			result.WriteString(word)
			lineLength += len(word)
		}
	}

	return result.String()
}

// escapeAppleScriptString escapes special characters in AppleScript strings
func escapeAppleScriptString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

// IsZenityInstalled checks if zenity is installed on Linux
func IsZenityInstalled() bool {
	if runtime.GOOS != "linux" {
		return true // On non-Linux systems, this check is irrelevant
	}

	_, err := exec.LookPath("zenity")
	return err == nil
}

// CheckDependencies checks if all required dependencies are installed
func CheckDependencies() error {
	switch runtime.GOOS {
	case "linux":
		if !IsZenityInstalled() {
			return fmt.Errorf("zenity is not installed. Please install it using your package manager")
		}
	case "darwin":
		// osascript is built into macOS, no need to check
		return nil
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
	return nil
}
