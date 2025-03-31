package server

import (
	"context"
	"testing"

	"github.com/nazar256/user-prompt-mcp/pkg/prompt"
)

// MockPromptService is a mock implementation of the prompt service
type MockPromptService struct {
	InputResponse string
	InputError    error
}

// PromptForInput is a mock implementation that returns predefined responses
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

func TestMCPServer_RegisterUserPromptTool(t *testing.T) {
	// Create a mock prompt service that satisfies the prompt.Service interface
	mockPrompt := &prompt.Service{}

	// Create an MCP server with the mock prompt service
	mcpServer := NewMCPServer(mockPrompt)

	// Register the user_prompt tool
	mcpServer.RegisterUserPromptTool()

	// We're just testing that the method runs without error
}
