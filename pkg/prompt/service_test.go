package prompt

import (
	"context"
	"testing"
	"time"
)

// MockDialogProvider implements the DialogProvider interface for testing
type MockDialogProvider struct {
	Response      string
	Error         error
	DelayDuration time.Duration
}

// ShowInputDialog implements the DialogProvider interface
func (m *MockDialogProvider) ShowInputDialog(prompt string, title string) (string, error) {
	// Simulate delay to test timeout
	if m.DelayDuration > 0 {
		time.Sleep(m.DelayDuration)
	}
	return m.Response, m.Error
}

func TestNewService(t *testing.T) {
	// Test with default options
	service := NewService(ServiceOptions{})
	if service == nil {
		t.Fatal("Failed to create service with default options")
	}

	// Test with custom options
	mockDialog := &MockDialogProvider{Response: "test", Error: nil}
	service = NewService(ServiceOptions{
		Dialog:     mockDialog,
		Timeout:    time.Second * 10,
		DefaultMsg: "Custom test message",
	})

	if service.dialog != mockDialog {
		t.Error("Dialog not set correctly")
	}
	if service.timeout != time.Second*10 {
		t.Error("Timeout not set correctly")
	}
	if service.defaultMsg != "Custom test message" {
		t.Error("Default message not set correctly")
	}
}

func TestPromptForInput(t *testing.T) {
	// Test successful prompt
	mockDialog := &MockDialogProvider{Response: "test response", Error: nil}
	service := NewService(ServiceOptions{Dialog: mockDialog})

	result, err := service.PromptForInput(context.Background(), PromptOptions{
		Prompt: "Test prompt",
		Title:  "Test title",
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if result != "test response" {
		t.Errorf("Expected result 'test response', got: %s", result)
	}

	// Test timeout
	mockDialog = &MockDialogProvider{
		Response:      "delayed response",
		Error:         nil,
		DelayDuration: time.Millisecond * 100,
	}
	service = NewService(ServiceOptions{
		Dialog:  mockDialog,
		Timeout: time.Millisecond * 50, // Less than the delay
	})

	_, err = service.PromptForInput(context.Background(), PromptOptions{})
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}
