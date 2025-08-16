package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ChatMessage represents a message in a chat conversation
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents a request to a chat completion API
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

// ChatResponse represents a response from a chat completion API
type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// HTTPClient interface for dependency injection
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// defaultHTTPClient is the default HTTP client
var defaultHTTPClient HTTPClient = &http.Client{
	Timeout: 30 * time.Second,
}

// getMockServerURL returns the mock server URL if in test mode
func getMockServerURL() string {
	if os.Getenv("GPT_CLI_TEST") == "1" {
		// Allow override via environment variable
		if mockURL := os.Getenv("MOCK_SERVER_URL"); mockURL != "" {
			return mockURL
		}
		return "http://127.0.0.1:8086"
	}
	return ""
}

// callOpenAIProvider calls the OpenAI API
func callOpenAIProvider(config *CoreConfig, opts *ProviderOptions) (*ProviderResponse, error) {
	baseURL := "https://api.openai.com"
	if opts.BaseURL != "" {
		baseURL = opts.BaseURL
	}
	if mockURL := getMockServerURL(); mockURL != "" {
		baseURL = mockURL
	}

	return callOpenAICompatibleAPI(baseURL, config, opts)
}

// callCopilotProvider calls the GitHub Copilot API
func callCopilotProvider(config *CoreConfig, opts *ProviderOptions) (*ProviderResponse, error) {
	baseURL := "https://api.github.com"
	if opts.BaseURL != "" {
		baseURL = opts.BaseURL
	}
	if mockURL := getMockServerURL(); mockURL != "" {
		baseURL = mockURL
	}

	return callOpenAICompatibleAPI(baseURL, config, opts)
}

// callGeminiProvider calls the Google Gemini API
func callGeminiProvider(config *CoreConfig, opts *ProviderOptions) (*ProviderResponse, error) {
	// For now, use the same OpenAI-compatible interface
	// TODO: Implement proper Gemini API calls
	baseURL := "https://generativelanguage.googleapis.com"
	if opts.BaseURL != "" {
		baseURL = opts.BaseURL
	}
	if mockURL := getMockServerURL(); mockURL != "" {
		baseURL = mockURL
	}

	return callOpenAICompatibleAPI(baseURL, config, opts)
}

// callOpenAICompatibleAPI makes a call to an OpenAI-compatible API
func callOpenAICompatibleAPI(baseURL string, config *CoreConfig, opts *ProviderOptions) (*ProviderResponse, error) {
	// Build chat request
	messages := []ChatMessage{}
	
	if config.System != "" {
		messages = append(messages, ChatMessage{
			Role:    "system",
			Content: config.System,
		})
	}
	
	messages = append(messages, ChatMessage{
		Role:    "user",
		Content: config.Prompt,
	})

	request := ChatRequest{
		Model:       config.Model,
		Messages:    messages,
		Temperature: config.Temperature,
		Stream:      false, // Non-streaming for now
	}

	// Marshal request to JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := baseURL + "/v1/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if opts.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+opts.APIKey)
	}

	// Make the request
	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for non-200 status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract content
	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	content := chatResp.Choices[0].Message.Content

	return &ProviderResponse{
		Text:     content,
		Markdown: content, // For now, treat content as both text and markdown
	}, nil
}