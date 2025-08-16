// Simple mock server for testing the CLI
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type MockServerTest struct{}

type MockChatRequest struct {
	Model       string `json:"model"`
	Messages    []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
	Temperature float64 `json:"temperature,omitempty"`
	Stream      bool    `json:"stream,omitempty"`
}

type MockChatResponse struct {
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
}

func mockHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MockChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var userMessage string
	for _, msg := range req.Messages {
		if msg.Role == "user" {
			userMessage = msg.Content
			break
		}
	}

	// Generate mock response
	responseContent := "Hello! How can I help you today?"
	if strings.Contains(strings.ToLower(userMessage), "markdown") {
		responseContent = "# Markdown Response\n\nThis is a **bold** response with *italic* text."
	} else if strings.Contains(strings.ToLower(userMessage), "test") {
		responseContent = "This is a test response from the mock server."
	}

	response := MockChatResponse{
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
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/chat/completions", mockHandler)

	log.Println("Starting mock server on :8086")
	log.Fatal(http.ListenAndServe(":8086", mux))
}