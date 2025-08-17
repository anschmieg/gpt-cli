package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/anschmieg/gpt-cli/internal/config"
)

// GeminiProvider implements the Provider interface for Google Gemini
type GeminiProvider struct {
	config *config.Config
	client *http.Client
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(config *config.Config) *GeminiProvider {
	return &GeminiProvider{
		config: config,
		client: &http.Client{},
	}
}

// GetName returns the provider name
func (p *GeminiProvider) GetName() string {
	return "gemini"
}

// CallProvider makes a request to Gemini API
func (p *GeminiProvider) CallProvider(prompt string) (string, error) {
	url := strings.TrimSuffix(p.config.BaseURL, "/") + "/chat/completions"

	requestBody := map[string]interface{}{
		"model": p.config.Model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": p.config.System,
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": p.config.Temperature,
		"stream":      false,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
			Code    string `json:"code"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if response.Error.Message != "" {
		return "", NewProviderError(response.Error.Message, response.Error.Code, nil)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return response.Choices[0].Message.Content, nil
}

// StreamProvider streams responses from Gemini API
func (p *GeminiProvider) StreamProvider(prompt string) (<-chan string, <-chan error) {
	contentChan := make(chan string)
	errorChan := make(chan error, 1)

	go func() {
		defer close(contentChan)
		defer close(errorChan)

		result, err := p.CallProvider(prompt)
		if err != nil {
			errorChan <- err
			return
		}

		// Simulate streaming by sending chunks
		words := strings.Fields(result)
		for _, word := range words {
			contentChan <- word + " "
		}
	}()

	return contentChan, errorChan
}