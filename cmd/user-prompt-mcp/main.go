package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/nazar256/user-prompt-mcp/internal/server"
	"github.com/nazar256/user-prompt-mcp/pkg/gui"
	"github.com/nazar256/user-prompt-mcp/pkg/prompt"
)

func main() {
	// Set up logging
	log.SetPrefix("[UserPromptMCP] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting User Prompt MCP Server...")

	// Parse command line flags
	timeoutSeconds := flag.Int("timeout", 0, "Timeout in seconds for user input (default: 1200)")
	flag.Parse()

	// Check environment variable for timeout
	if envTimeout := os.Getenv("USER_PROMPT_TIMEOUT"); envTimeout != "" {
		if seconds, err := strconv.Atoi(envTimeout); err == nil {
			timeoutSeconds = &seconds
		} else {
			log.Printf("Warning: Invalid USER_PROMPT_TIMEOUT value: %s", envTimeout)
		}
	}

	// Check for required dependencies
	if err := gui.CheckDependencies(); err != nil {
		log.Fatalf("Dependency check failed: %v", err)
	}

	// Create prompt service options
	opts := prompt.DefaultOptions()
	if *timeoutSeconds > 0 {
		opts.Timeout = time.Duration(*timeoutSeconds) * time.Second
	}

	// Create a prompt service with configured options
	promptService := prompt.NewService(opts)
	log.Printf("Prompt service initialized with timeout: %v", opts.Timeout)

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
