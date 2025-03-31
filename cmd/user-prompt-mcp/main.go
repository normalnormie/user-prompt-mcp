package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nazar256/user-prompt-mcp/internal/server"
	"github.com/nazar256/user-prompt-mcp/pkg/gui"
	"github.com/nazar256/user-prompt-mcp/pkg/prompt"
)

func main() {
	// Set up logging
	log.SetPrefix("[UserPromptMCP] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting User Prompt MCP Server...")

	// Check for required dependencies
	if err := gui.CheckDependencies(); err != nil {
		log.Fatalf("Dependency check failed: %v", err)
	}

	// Create a prompt service with default options
	promptService := prompt.NewService(prompt.DefaultOptions())
	log.Println("Prompt service initialized")

	// Create MCP server
	mcpServer := server.NewMCPServer(promptService)

	// Register the user_prompt tool
	mcpServer.RegisterUserPromptTool()

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Handle signals in a separate goroutine
	go func() {
		<-sigCh
		log.Println("Received shutdown signal. Shutting down...")
		os.Exit(0)
	}()

	// Start the server
	log.Println("Server starting. Waiting for requests...")
	if err := mcpServer.ServeStdio(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
