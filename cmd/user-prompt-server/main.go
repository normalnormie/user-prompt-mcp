package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const httpPort = "3030"

// --- Structures to manage the active prompt for Vibeframe ---
type activePrompt struct {
	Prompt       string
	Title        string
	ResponseChan chan string // Channel to send the user's response back
	ErrorChan    chan error  // Channel to send an error
	IsActive     bool
}

var currentPrompt struct {
	sync.Mutex
	details *activePrompt
}

// --- End of Vibeframe prompt state ---

// --- HTTP Handlers for Vibeframe UI ---
func vibeframeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("HTTP: Received request for /vibeframe")
	htmlContent := `
<!DOCTYPE html>
<html>
<head>
    <title>User Prompt</title>
    <style>
        body { font-family: sans-serif; margin: 20px; background-color: #2e2e2e; color: #d4d4d4; }
        .container { max-width: 500px; margin: auto; padding: 20px; background-color: #3c3c3c; border-radius: 8px; box-shadow: 0 0 10px rgba(0,0,0,0.5); }
        h2 { color: #569cd6; }
        label { display: block; margin-bottom: 8px; }
        input[type="text"], textarea { width: calc(100% - 22px); padding: 10px; margin-bottom: 20px; border-radius: 4px; border: 1px solid #555; background-color: #252526; color: #d4d4d4; box-sizing: border-box; }
        textarea { min-height: 80px; }
        button { padding: 10px 15px; border: none; border-radius: 4px; background-color: #0e639c; color: white; cursor: pointer; }
        button:hover { background-color: #1177bb; }
        #promptText { margin-bottom: 15px; white-space: pre-wrap; }
        #inputForm { display: none; } /* Hidden initially */
    </style>
</head>
<body>
    <div class="container">
        <h2 id="promptTitle"></h2>
        <p id="promptText">Waiting for LLM prompt...</p>
        <form id="inputForm">
            <label for="userInput">Your input:</label>
            <textarea id="userInput" name="userInput" required></textarea>
            <button type="submit">Submit</button>
        </form>
    </div>
    <script>
        const promptTitleElement = document.getElementById('promptTitle');
        const promptTextElement = document.getElementById('promptText');
        const inputForm = document.getElementById('inputForm');
        const userInputElement = document.getElementById('userInput');

        userInputElement.addEventListener('keydown', function(event) {
            if (event.key === 'Enter' && !event.shiftKey) {
                event.preventDefault(); // Prevent new line
                // Find the submit button within the form and click it
                const submitButton = inputForm.querySelector('button[type="submit"]');
                if (submitButton) {
                    submitButton.click();
                }
            }
        });

        const eventSource = new EventSource('/events');
        eventSource.onmessage = function(event) {
            const data = JSON.parse(event.data);
            if (data.type === 'prompt') {
                promptTitleElement.textContent = data.title || 'User Input Required';
                promptTextElement.textContent = data.prompt || 'Please provide input:';
                userInputElement.value = '';
                inputForm.style.display = 'block';
                userInputElement.focus();
            } else if (data.type === 'close') {
                promptTitleElement.textContent = 'Prompt Closed';
                promptTextElement.textContent = 'The prompt has been closed or timed out by the server: ' + (data.reason || '');
                inputForm.style.display = 'none';
                // Do not close eventSource here, allow server to re-prompt if needed later
            }
        };
        eventSource.onerror = function(err) {
            console.error("EventSource failed:", err);
            promptTextElement.textContent = "Error connecting to prompt server. Please try reloading Vibeframe or ensure the prompt server is running.";
            // Consider not closing eventSource to allow auto-reconnect if server comes back
        };

        inputForm.addEventListener('submit', function(e) {
            e.preventDefault();
            const input = userInputElement.value;
            fetch('/submit-input', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ input: input })
            })
            .then(response => {
                if (!response.ok) {
                    response.text().then(text => {
                        promptTextElement.textContent = "Input submission failed: " + text;
                        // Keep form visible for retry or show error message
                    });
                } else {
                     promptTitleElement.textContent = "Input submitted.";
                     promptTextElement.textContent = "Waiting for processing...";
                     // The main application (client) will close the prompt or issue a new one.
                }
                // Do not hide form immediately, server response or new prompt will dictate UI
            })
            .catch(error => {
                console.error('Error submitting input:', error);
                promptTextElement.textContent = "Error submitting input: " + error;
            });
        });
    </script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; connect-src 'self';")
	w.Write([]byte(htmlContent))
}

var sseClients sync.Map // map[string]chan []byte, key is client remote addr or unique ID

func eventsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("HTTP: Client connected to /events (SSE)")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	messageChan := make(chan []byte, 10)
	clientKey := r.RemoteAddr // Consider a more unique ID if needed
	sseClients.Store(clientKey, messageChan)
	log.Printf("HTTP: SSE client %s registered", clientKey)

	// Send current prompt if one is active
	currentPrompt.Lock()
	if currentPrompt.details != nil && currentPrompt.details.IsActive {
		promptData := fmt.Sprintf(`{"type": "prompt", "prompt": %q, "title": %q}`, currentPrompt.details.Prompt, currentPrompt.details.Title)
		log.Printf("HTTP: SSE client %s - Sending initial active prompt: %s", clientKey, promptData)
		fmt.Fprintf(w, "data: %s\n\n", promptData)
		flusher.Flush()
	}
	currentPrompt.Unlock()

	defer func() {
		log.Printf("HTTP: SSE client %s - DEFER function in eventsHandler started.", clientKey)
		sseClients.Delete(clientKey)
		close(messageChan)
		log.Printf("HTTP: SSE client %s disconnected and cleaned up. messageChan closed.", clientKey)
	}()

	// Keep connection open and send messages
	log.Printf("HTTP: SSE client %s - Entering message loop.", clientKey)
	for {
		select {
		case msg, ok := <-messageChan:
			if !ok { // Channel closed
				log.Printf("HTTP: SSE client %s - messageChan closed by sender. Terminating handler.", clientKey)
				return // Exit handler, which triggers defer
			}
			log.Printf("HTTP: SSE client %s - Sending message: %s", clientKey, string(msg))
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done(): // Client disconnected OR server shutting down connection
			log.Printf("HTTP: SSE client %s - r.Context().Done() signaled. Error: %v. Terminating handler.", clientKey, r.Context().Err())
			return // Exit handler, which triggers defer
		}
	}
}

func submitInputHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("HTTP: Received request for /submit-input")
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*") // For webview

	var data struct {
		Input string `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Printf("HTTP: Error decoding /submit-input JSON: %v", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	currentPrompt.Lock()
	defer currentPrompt.Unlock()

	if currentPrompt.details != nil && currentPrompt.details.IsActive {
		log.Printf("HTTP: Received input %q for active prompt", data.Input)
		currentPrompt.details.ResponseChan <- data.Input
		// IsActive will be set to false by the /api/trigger-prompt handler once it receives the response.
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Input received by server."))
	} else {
		log.Println("HTTP: Received input via POST, but no active prompt or prompt already handled.")
		http.Error(w, "No active prompt or prompt already handled", http.StatusConflict)
	}
}

func broadcastSSEMessage(message []byte) {
	log.Printf("HTTP: Broadcasting SSE message: %s", string(message))
	sseClients.Range(func(key, value interface{}) bool {
		clientChan, ok := value.(chan []byte)
		if ok {
			select {
			case clientChan <- message:
			default:
				log.Printf("HTTP: SSE client channel for %v is full, skipping broadcast.", key)
			}
		}
		return true
	})
}

// --- API Handler for triggering prompts ---
type TriggerPromptRequest struct {
	Prompt    string `json:"prompt"`
	Title     string `json:"title"`
	TimeoutMs int64  `json:"timeout_ms"`
}

type TriggerPromptResponse struct {
	Input string `json:"input,omitempty"`
	Error string `json:"error,omitempty"`
}

func triggerPromptHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("API: Received request for /api/trigger-prompt")
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TriggerPromptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("API: Error decoding /api/trigger-prompt JSON: %v", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	log.Printf("API: Prompt request: Title=%q, Prompt=%q, Timeout=%dms", req.Title, req.Prompt, req.TimeoutMs)

	currentPrompt.Lock()
	if currentPrompt.details != nil && currentPrompt.details.IsActive {
		currentPrompt.Unlock()
		log.Println("API: Another prompt is already active.")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(TriggerPromptResponse{Error: "Another prompt is already active"})
		return
	}

	responseChan := make(chan string)
	errorChan := make(chan error)
	currentPrompt.details = &activePrompt{
		Prompt:       req.Prompt,
		Title:        req.Title,
		ResponseChan: responseChan,
		ErrorChan:    errorChan,
		IsActive:     true,
	}
	currentPrompt.Unlock() // Unlock before broadcasting and waiting

	promptData := fmt.Sprintf(`{"type": "prompt", "prompt": %q, "title": %q}`, req.Prompt, req.Title)
	broadcastSSEMessage([]byte(promptData))

	var timeoutDuration time.Duration
	if req.TimeoutMs > 0 {
		timeoutDuration = time.Duration(req.TimeoutMs) * time.Millisecond
	} else {
		timeoutDuration = 20 * time.Minute // Default fallback timeout
	}

	var resp TriggerPromptResponse
	select {
	case input := <-responseChan:
		log.Printf("API: Received input from Vibeframe: %q", input)
		resp.Input = input
		w.WriteHeader(http.StatusOK)
	case err := <-errorChan: // This channel is not currently written to by submitInput, but could be for other errors
		log.Printf("API: Error channel signaled: %v", err)
		resp.Error = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	case <-time.After(timeoutDuration):
		log.Printf("API: Prompt timed out after %v", timeoutDuration)
		resp.Error = "Prompt timed out"
		w.WriteHeader(http.StatusGatewayTimeout) // Or another appropriate error
		// Ensure prompt is marked inactive
		currentPrompt.Lock()
		if currentPrompt.details != nil && currentPrompt.details.IsActive { // Check if it's the same prompt
			currentPrompt.details.IsActive = false
			broadcastSSEMessage([]byte(fmt.Sprintf(`{"type": "close", "reason": %q}`, "timeout")))
		}
		currentPrompt.Unlock()
	}

	// Mark prompt as inactive after handling, regardless of outcome
	currentPrompt.Lock()
	if currentPrompt.details != nil { // Could have been cleared by timeout already
		currentPrompt.details.IsActive = false
	}
	currentPrompt.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	log.SetPrefix("[UserPromptServer] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	log.Println("----------------------------------------------------")
	log.Println("Starting User Prompt Server (Vibeframe HTTP/S Server)...")

	port := flag.String("port", httpPort, "Port for the HTTP/S server")
	tlsCertFile := flag.String("tls-cert-file", "", "Path to TLS certificate file (for HTTPS)")
	tlsKeyFile := flag.String("tls-key-file", "", "Path to TLS key file (for HTTPS)")
	flag.Parse()

	http.HandleFunc("/vibeframe", vibeframeHandler)
	http.HandleFunc("/events", eventsHandler)
	http.HandleFunc("/submit-input", submitInputHandler)
	http.HandleFunc("/api/trigger-prompt", triggerPromptHandler)

	serverAddr := ":" + *port
	server := &http.Server{Addr: serverAddr}

	go func() {
		var serverErr error
		if *tlsCertFile != "" && *tlsKeyFile != "" {
			log.Printf("Vibeframe HTTPS server starting on port %s", *port)
			serverErr = server.ListenAndServeTLS(*tlsCertFile, *tlsKeyFile)
		} else {
			log.Printf("Vibeframe HTTP server starting on port %s", *port)
			serverErr = server.ListenAndServe()
		}
		if serverErr != nil && serverErr != http.ErrServerClosed {
			log.Fatalf("Server ListenAndServe/ListenAndServeTLS error: %v", serverErr)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutdown signal received, gracefully shutting down server...")

	log.Println("Closing active SSE client channels...")
	sseClients.Range(func(key, value interface{}) bool {
		// The variable 'clientChan' (the channel itself) is intentionally not used directly here
		// to avoid potential double-close panics. The primary mechanism for closing
		// the SSE handler and its channel is via r.Context().Done() being triggered
		// by server.Shutdown(), and then the 'defer func()' in eventsHandler
		// closes its specific messageChan. This loop is more for acknowledging
		// the step or for future, safer signaling mechanisms if needed.
		return true
	})
	// The http.Server.Shutdown method should handle closing client connections,
	// which in turn should cancel the request's context (r.Context()).
	// The eventsHandler is designed to exit when r.Context() is done.

	// Let's ensure the shutdown timeout is clear.
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelShutdown()

	log.Println("Attempting http.Server.Shutdown()...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP server Shutdown error: %v", err)
	}
	log.Println("Server gracefully stopped.")
}
