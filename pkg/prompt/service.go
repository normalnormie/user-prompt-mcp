package prompt

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nazar256/user-prompt-mcp/pkg/gui"
)

const defaultPromptServerURL = "http://localhost:3030"

// Service handles user input prompts
type Service struct {
	dialog     gui.DialogProvider
	timeout    time.Duration
	defaultMsg string
	mutex      sync.Mutex
}

// ServiceOptions contains options for creating a new PromptService
type ServiceOptions struct {
	Dialog     gui.DialogProvider
	Timeout    time.Duration
	DefaultMsg string
}

// DefaultOptions returns the default options for the prompt service
func DefaultOptions() ServiceOptions {
	return ServiceOptions{
		Dialog:     gui.NewRemoteDialog(defaultPromptServerURL),
		Timeout:    time.Minute * 20, // 20 minute default timeout
		DefaultMsg: "Cursor is requesting additional input",
	}
}

// NewService creates a new prompt service with the given options
func NewService(opts ServiceOptions) *Service {
	if opts.Dialog == nil {
		opts.Dialog = gui.NewRemoteDialog(defaultPromptServerURL)
	}
	if opts.Timeout == 0 {
		opts.Timeout = DefaultOptions().Timeout
	}
	if opts.DefaultMsg == "" {
		opts.DefaultMsg = DefaultOptions().DefaultMsg
	}

	return &Service{
		dialog:     opts.Dialog,
		timeout:    opts.Timeout,
		defaultMsg: opts.DefaultMsg,
	}
}

// PromptOptions contains options for a specific prompt
type PromptOptions struct {
	Prompt     string
	Title      string
	Timeout    time.Duration
	DefaultMsg string
}

// PromptForInput displays a prompt to the user and returns their input
// The prompt is displayed with the specified options and will timeout after the specified duration
func (s *Service) PromptForInput(ctx context.Context, opts PromptOptions) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Use default values if not provided
	if opts.Title == "" {
		opts.Title = "User Input Required"
	}
	if opts.Prompt == "" {
		opts.Prompt = s.defaultMsg
	}
	if opts.Timeout == 0 {
		opts.Timeout = s.timeout
	}

	// Create a timeout context based on the provided context and the prompt timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	// Channel to receive result
	resultCh := make(chan struct {
		result string
		err    error
	})

	// Run the dialog in a goroutine
	go func() {
		result, err := s.dialog.ShowInputDialog(timeoutCtx, opts.Prompt, opts.Title)
		resultCh <- struct {
			result string
			err    error
		}{result, err}
	}()

	// Wait for result or timeout
	select {
	case response := <-resultCh:
		if response.err != nil {
			return "", response.err
		}
		return response.result, nil
	case <-timeoutCtx.Done():
		// Check if the original context was cancelled or if it was our timeout
		if ctx.Err() != nil {
			return "", fmt.Errorf("prompt cancelled: %w", ctx.Err())
		}
		return "", fmt.Errorf("prompt timed out after %v", opts.Timeout)
	}
}
