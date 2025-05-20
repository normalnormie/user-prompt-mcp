package gui

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// RemoteDialog implements DialogProvider by making HTTP calls to a separate server.
type RemoteDialog struct {
	ServerURL string // e.g., "http://localhost:3030"
	Client    *http.Client
}

// NewRemoteDialog creates a new RemoteDialog.
// serverURL should be the base URL of the user-prompt-server (e.g., "http://localhost:3030").
func NewRemoteDialog(serverURL string) *RemoteDialog {
	return &RemoteDialog{
		ServerURL: serverURL,
		Client: &http.Client{
			Timeout: 0, // Context will handle overall timeout for the request
		},
	}
}

type TriggerPromptRequest struct {
	Prompt    string `json:"prompt"`
	Title     string `json:"title"`
	TimeoutMs int64  `json:"timeout_ms"`
}

type TriggerPromptResponse struct {
	Input string `json:"input,omitempty"`
	Error string `json:"error,omitempty"`
}

// ShowInputDialog sends a prompt request to the remote server and waits for the response.
func (rd *RemoteDialog) ShowInputDialog(ctx context.Context, prompt string, title string) (string, error) {
	var timeoutMs int64
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining > 0 {
			timeoutMs = remaining.Milliseconds()
		} else {
			// Context already expired or very close to it
			return "", fmt.Errorf("prompt context already expired before calling remote server: %w", context.DeadlineExceeded)
		}
	} else {
		timeoutMs = (20 * time.Minute).Milliseconds() // Default if no deadline on context
		log.Println("Warning: No deadline on context for RemoteDialog.ShowInputDialog, using default timeout for server call.")
	}

	if timeoutMs <= 0 { // Ensure we don't send a non-positive timeout
		return "", fmt.Errorf("prompt context resulted in non-positive timeout: %dms", timeoutMs)
	}

	requestPayload := TriggerPromptRequest{
		Prompt:    prompt,
		Title:     title,
		TimeoutMs: timeoutMs,
	}

	payloadBytes, err := json.Marshal(requestPayload)
	if err != nil {
		log.Printf("RemoteDialog: Error marshalling request: %v", err)
		return "", fmt.Errorf("failed to marshal prompt request: %w", err)
	}

	reqURL := rd.ServerURL + "/api/trigger-prompt"
	log.Printf("RemoteDialog: Sending prompt request to %s with timeout %dms", reqURL, timeoutMs)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("RemoteDialog: Error creating HTTP request: %v", err)
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := rd.Client.Do(httpReq)
	if err != nil {
		log.Printf("RemoteDialog: Error sending HTTP request to server: %v", err)
		// Check if context error is the cause
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return "", fmt.Errorf("prompt request to server timed out or was cancelled: %w", err)
		}
		return "", fmt.Errorf("failed to send prompt request to server: %w", err)
	}
	defer httpResp.Body.Close()

	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		log.Printf("RemoteDialog: Error reading response body: %v", err)
		return "", fmt.Errorf("failed to read response from server: %w", err)
	}

	log.Printf("RemoteDialog: Received response from server: Status=%s, Body=%s", httpResp.Status, string(bodyBytes))

	var serverResponse TriggerPromptResponse
	if err := json.Unmarshal(bodyBytes, &serverResponse); err != nil {
		log.Printf("RemoteDialog: Error unmarshalling server response: %v. Body: %s", err, string(bodyBytes))
		// If unmarshalling fails, but status was OK, it's an issue.
		// If status was not OK, the error might be in plain text or non-JSON.
		if httpResp.StatusCode == http.StatusOK {
			return "", fmt.Errorf("failed to unmarshal server response: %w (body: %s)", err, string(bodyBytes))
		}
		// For non-OK status, try to return a generic error with status and body
		return "", fmt.Errorf("server returned error: status %s, body: %s", httpResp.Status, string(bodyBytes))
	}

	if httpResp.StatusCode != http.StatusOK {
		if serverResponse.Error != "" {
			return "", fmt.Errorf("server error: %s (status %s)", serverResponse.Error, httpResp.Status)
		}
		return "", fmt.Errorf("server returned non-OK status: %s, with body: %s", httpResp.Status, string(bodyBytes))
	}

	if serverResponse.Error != "" {
		// This case should ideally be covered by non-OK status, but good to check
		return "", fmt.Errorf("prompt failed on server: %s", serverResponse.Error)
	}

	return serverResponse.Input, nil
}

// CheckDependencies for RemoteDialog - none needed as it's network-based.
func (rd *RemoteDialog) CheckDependencies() error {
	// Could add a ping to the server here if desired
	return nil
}
