package main

import (
	"flag"
	"log"
	"time"

	"github.com/nazar256/user-prompt-mcp/internal/server"
	"github.com/nazar256/user-prompt-mcp/pkg/gui"
	"github.com/nazar256/user-prompt-mcp/pkg/prompt"
)

const defaultPromptServerURL = "http://localhost:3030"

func main() {
	log.SetPrefix("[UserPromptClient] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	log.Println("----------------------------------------------------")
	log.Println("Starting User Prompt MCP Client...")

	timeoutSeconds := flag.Int("timeout", 0, "Default timeout in seconds for user input (default: 1200 from prompt.Service)")
	promptServerURL := flag.String("prompt-server-url", defaultPromptServerURL, "URL of the user-prompt-server")
	flag.Parse()

	opts := prompt.DefaultOptions()
	if *timeoutSeconds > 0 {
		opts.Timeout = time.Duration(*timeoutSeconds) * time.Second
	}

	log.Printf("Configuring to use remote prompt server at: %s", *promptServerURL)
	opts.Dialog = gui.NewRemoteDialog(*promptServerURL)

	if err := opts.Dialog.CheckDependencies(); err != nil {
		log.Fatalf("Remote dialog dependency check failed: %v", err)
	}
	log.Println("Using RemoteDialog provider.")

	promptService := prompt.NewService(opts)
	log.Printf("Prompt service initialized with default timeout: %v", opts.Timeout)

	mcpServer := server.NewMCPServer(promptService)
	mcpServer.RegisterUserPromptTool()

	log.Println("MCP Client (stdio server) starting. Waiting for stdio requests from Cursor...")
	if err := mcpServer.ServeStdio(); err != nil {
		log.Fatalf("MCP Client (stdio server) error: %v", err)
	}
	log.Println("MCP Client (stdio server) finished.")
}
