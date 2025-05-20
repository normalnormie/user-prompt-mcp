package gui

import (
	"context"
	"errors"
	"log"
)

// VibeframeDialog implements DialogProvider for Vibeframe integration.
// It doesn't directly show a dialog but signals the main HTTP server part
// to make a prompt available for Vibeframe clients.
type VibeframeDialog struct {
	// Function to call to request a prompt and wait for its result.
	// This function is provided by main.go and interacts with the global prompt state.
	requestPromptFunc func(ctx context.Context, prompt, title string) (string, error)
}

// NewVibeframeDialog creates a new VibeframeDialog.
// The `requestPromptFunc` is the core logic that will:
// 1. Set the global `currentPrompt`.
// 2. Notify connected Vibeframe clients (e.g., via SSE).
// 3. Wait on channels for the user's input (from an HTTP handler) or a timeout from context.
func NewVibeframeDialog(promptRequester func(ctx context.Context, prompt, title string) (string, error)) *VibeframeDialog {
	if promptRequester == nil {
		// This should not happen if initialized correctly from main.go
		log.Fatal("VibeframeDialog: promptRequester function cannot be nil")
		return nil // Or handle error appropriately
	}
	return &VibeframeDialog{
		requestPromptFunc: promptRequester,
	}
}

// ShowInputDialog for VibeframeDialog uses the provided requestPromptFunc.
func (v *VibeframeDialog) ShowInputDialog(ctx context.Context, prompt string, title string) (string, error) {
	if v.requestPromptFunc == nil {
		log.Println("Error: VibeframeDialog.requestPromptFunc is not set.")
		return "", errors.New("VibeframeDialog not properly initialized")
	}
	// Pass the context, prompt, and title to the actual prompting logic
	return v.requestPromptFunc(ctx, prompt, title)
}

// CheckDependencies for VibeframeDialog - none needed as it's web-based.
func (v *VibeframeDialog) CheckDependencies() error {
	return nil // No external CLI dependencies like zenity
}
