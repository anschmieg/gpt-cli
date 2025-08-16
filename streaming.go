package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// StreamingResponse represents a chunk from a streaming response
type StreamingResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

// streamChatCompletion handles streaming chat completions
func streamChatCompletion(config *CoreConfig, opts *ProviderOptions) error {
	baseURL := getProviderBaseURL(config.Provider, opts)
	
	// Build the request
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
		Stream:      true,
	}

	// Marshal request
	requestBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := baseURL + "/v1/chat/completions"
	req, err := http.NewRequest("POST", url, strings.NewReader(string(requestBody)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	if opts.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+opts.APIKey)
	}

	// Make the request
	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Initialize streaming markdown renderer if markdown is enabled
	var streamRenderer *StreamingMarkdownRenderer
	if config.UseMarkdown {
		streamRenderer = NewStreamingMarkdownRenderer(true)
	}

	// Process streaming response
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// Parse SSE data
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			
			// Check for end of stream
			if data == "[DONE]" {
				break
			}

			// Parse the JSON chunk
			var chunk StreamingResponse
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				if config.Verbose {
					fmt.Fprintf(os.Stderr, "Failed to parse chunk: %v\n", err)
				}
				continue
			}

			// Output the content
			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				content := chunk.Choices[0].Delta.Content
				
				if config.UseMarkdown && streamRenderer != nil {
					// Process through streaming markdown renderer
					rendered := streamRenderer.ProcessChunk(content)
					if rendered != "" {
						fmt.Print(rendered)
					}
				} else {
					// Output raw content
					fmt.Print(content)
				}
			}
		}
	}

	// Flush any remaining content from the streaming renderer
	if config.UseMarkdown && streamRenderer != nil {
		if remaining := streamRenderer.Flush(); remaining != "" {
			fmt.Print(remaining)
		}
	}

	// Add final newline
	fmt.Println()
	
	return scanner.Err()
}

// getProviderBaseURL returns the base URL for a provider
func getProviderBaseURL(provider string, opts *ProviderOptions) string {
	if opts.BaseURL != "" {
		return opts.BaseURL
	}
	
	if mockURL := getMockServerURL(); mockURL != "" {
		return mockURL
	}

	provider = strings.ToLower(provider)
	switch provider {
	case "openai":
		return "https://api.openai.com"
	case "copilot":
		return "https://api.github.com"
	case "gemini":
		return "https://generativelanguage.googleapis.com"
	default:
		return "https://api.openai.com"
	}
}

// updateTryStreamingProvider to use the new streaming implementation
func tryStreamingProvider(config *CoreConfig, providerOpts *ProviderOptions) error {
	return streamChatCompletion(config, providerOpts)
}