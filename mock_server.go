package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// MockServer provides a simple mock OpenAI-compatible API for testing
type MockServer struct {
	server *http.Server
}

// ChatCompletionResponse represents the structure of a chat completion response
type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// NewMockServer creates a new mock server
func NewMockServer(port string) *MockServer {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/chat/completions", handleChatCompletions)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	return &MockServer{server: server}
}

// Start starts the mock server
func (m *MockServer) Start() error {
	log.Printf("Starting mock server on %s", m.server.Addr)
	return m.server.ListenAndServe()
}

// Stop stops the mock server
func (m *MockServer) Stop() error {
	return m.server.Close()
}

// handleChatCompletions handles chat completions requests
func handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Extract the user message
	var userMessage string
	for _, msg := range req.Messages {
		if msg.Role == "user" {
			userMessage = msg.Content
			break
		}
	}

	// Generate a mock response based on the prompt
	responseContent := generateMockResponse(userMessage)

	// Create response
	response := ChatCompletionResponse{
		ID:      "chatcmpl-mock123",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   req.Model,
		Choices: []struct {
			Index   int `json:"index"`
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		}{
			{
				Index: 0,
				Message: struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{
					Role:    "assistant",
					Content: responseContent,
				},
				FinishReason: "stop",
			},
		},
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// generateMockResponse generates a mock response based on the input
func generateMockResponse(prompt string) string {
	prompt = strings.ToLower(prompt)
	
	if strings.Contains(prompt, "hello") {
		return "Hello! How can I help you today?"
	}
	if strings.Contains(prompt, "test") {
		return "This is a test response from the mock server."
	}
	if strings.Contains(prompt, "markdown") {
		return "# Markdown Response\n\nThis is a **bold** response with *italic* text and `code`."
	}
	
	return "This is a mock response to: " + prompt
}