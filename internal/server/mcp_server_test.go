package server

import (
	"context"
	"testing"

	"github.com/nazar256/user-input-mcp/pkg/prompt"
)

// MockPromptService mocks the prompt.Service for testing
type MockPromptService struct {
	InputResponse string
	InputError    error
}

// PromptForInput is a mock implementation of the PromptForInput method
func (m *MockPromptService) PromptForInput(ctx context.Context, opts prompt.PromptOptions) (string, error) {
	return m.InputResponse, m.InputError
}

func TestNewMCPServer(t *testing.T) {
	// Create a mock prompt service
	mockService := &prompt.Service{}

	// Create a new MCP server
	server := NewMCPServer(mockService)
	if server == nil {
		t.Fatal("Failed to create MCP server")
	}
	if server.mcpServer == nil {
		t.Error("MCP server not initialized")
	}
}

func TestRegisterUserPromptTool(t *testing.T) {
	// Create a prompt service
	mockService := &prompt.Service{}

	// Create a new MCP server
	server := NewMCPServer(mockService)

	// Register the user prompt tool
	server.RegisterUserPromptTool()

	// Unfortunately we can't directly access the tools list in the MCP server
	// So we'll just verify that the method doesn't panic
}
