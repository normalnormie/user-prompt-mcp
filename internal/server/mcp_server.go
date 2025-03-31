package server

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nazar256/user-prompt-mcp/pkg/prompt"
)

const (
	// ServerName is the name of the MCP server
	ServerName = "User Prompt MCP"
	// ServerVersion is the version of the MCP server
	ServerVersion = "1.0.0"
	// UserPromptToolName is the name of the user prompt tool
	UserPromptToolName = "user_prompt"
)

// MCPServer represents the MCP server for user input
type MCPServer struct {
	promptService *prompt.Service
	mcpServer     *server.MCPServer
}

// NewMCPServer creates a new MCP Server for user input
func NewMCPServer(promptService *prompt.Service) *MCPServer {
	// Create the MCP server with hooks for debugging
	hooks := &server.Hooks{}

	hooks.AddBeforeAny(func(id any, method mcp.MCPMethod, message any) {
		log.Printf("Request: [%v] %s", id, method)
	})

	hooks.AddOnSuccess(func(id any, method mcp.MCPMethod, message any, result any) {
		log.Printf("Success: [%v] %s", id, method)
	})

	hooks.AddOnError(func(id any, method mcp.MCPMethod, message any, err error) {
		log.Printf("Error: [%v] %s: %v", id, method, err)
	})

	// Create the server with error logging enabled
	mcpServer := server.NewMCPServer(
		ServerName,
		ServerVersion,
		server.WithLogging(),
		server.WithHooks(hooks),
	)

	return &MCPServer{
		promptService: promptService,
		mcpServer:     mcpServer,
	}
}

// RegisterUserPromptTool registers the user prompt tool with the MCP server
func (s *MCPServer) RegisterUserPromptTool() {
	// Create the user_prompt tool definition
	tool := mcp.NewTool(
		UserPromptToolName,
		mcp.WithDescription("Request additional input from the user during generation"),
		mcp.WithString("prompt",
			mcp.Description("The prompt to display to the user"),
			mcp.Required(),
		),
		mcp.WithString("title",
			mcp.Description("The title of the dialog window (optional)"),
		),
	)

	// Register the tool handler
	s.mcpServer.AddTool(tool, s.userPromptHandler)
	log.Printf("Registered tool: %s", UserPromptToolName)
}

// userPromptHandler handles calls to the user_prompt tool
func (s *MCPServer) userPromptHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract the prompt and title from the request
	promptText, ok := request.Params.Arguments["prompt"].(string)
	if !ok {
		return nil, errors.New("prompt argument must be a string")
	}

	// Title is optional
	var title string
	if titleArg, exists := request.Params.Arguments["title"]; exists {
		if titleStr, ok := titleArg.(string); ok {
			title = titleStr
		}
	}

	log.Printf("User prompt request: prompt=%q, title=%q", promptText, title)

	// Display the prompt to the user and get their input
	userInput, err := s.promptService.PromptForInput(ctx, prompt.PromptOptions{
		Prompt: promptText,
		Title:  title,
	})

	if err != nil {
		log.Printf("Error getting user input: %v", err)
		return nil, fmt.Errorf("failed to get user input: %w", err)
	}

	log.Printf("User provided input: %q", userInput)

	// Return the user's input
	return mcp.NewToolResultText(userInput), nil
}

// GetMCPServer returns the underlying MCP server
func (s *MCPServer) GetMCPServer() *server.MCPServer {
	return s.mcpServer
}

// ServeStdio runs the server using the stdio transport
func (s *MCPServer) ServeStdio() error {
	log.Println("Starting MCP server using stdio transport")
	return server.ServeStdio(s.mcpServer)
}
