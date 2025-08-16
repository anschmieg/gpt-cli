package main

import (
	"os"
	"testing"
	"time"
)

func TestIntegrationWithMockServer(t *testing.T) {
	// Set test environment
	os.Setenv("GPT_CLI_TEST", "1")
	defer os.Unsetenv("GPT_CLI_TEST")

	// Start mock server
	mockServer := NewMockServer("8086")
	go func() {
		if err := mockServer.Start(); err != nil && err.Error() != "http: Server closed" {
			t.Logf("Mock server error: %v", err)
		}
	}()
	
	// Give server time to start
	time.Sleep(100 * time.Millisecond)
	defer mockServer.Stop()

	tests := []struct {
		name   string
		config *CoreConfig
		opts   *ProviderOptions
	}{
		{
			name: "basic openai call",
			config: &CoreConfig{
				Provider:    "openai",
				Model:       "gpt-3.5-turbo",
				Temperature: 0.7,
				Prompt:      "hello world",
				UseMarkdown: true,
			},
			opts: &ProviderOptions{
				APIKey:  "test-key",
				BaseURL: "http://127.0.0.1:8086",
			},
		},
		{
			name: "copilot provider",
			config: &CoreConfig{
				Provider:    "copilot",
				Model:       "gpt-4o-mini",
				Temperature: 0.6,
				Prompt:      "test prompt",
				UseMarkdown: false,
			},
			opts: &ProviderOptions{
				APIKey:  "test-key",
				BaseURL: "http://127.0.0.1:8086",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runCore(tt.config, tt.opts)
			if err != nil {
				t.Errorf("runCore() error = %v", err)
			}
		})
	}
}

func TestProviderResponse(t *testing.T) {
	// Set test environment
	os.Setenv("GPT_CLI_TEST", "1")
	os.Setenv("MOCK_SERVER_URL", "http://127.0.0.1:8087")
	defer func() {
		os.Unsetenv("GPT_CLI_TEST")
		os.Unsetenv("MOCK_SERVER_URL")
	}()

	// Start mock server
	mockServer := NewMockServer("8087")
	go func() {
		mockServer.Start()
	}()
	
	// Give server time to start
	time.Sleep(100 * time.Millisecond)
	defer mockServer.Stop()

	config := &CoreConfig{
		Provider:    "openai",
		Model:       "gpt-3.5-turbo",
		Temperature: 0.7,
		Prompt:      "hello",
		UseMarkdown: true,
	}
	
	opts := &ProviderOptions{
		APIKey: "test-key",
		// Don't set BaseURL here, let it use the mock server URL from environment
	}

	response, err := callProvider(config, opts)
	if err != nil {
		t.Fatalf("callProvider() error = %v", err)
	}

	if response.Text == "" && response.Markdown == "" {
		t.Error("Expected non-empty response")
	}

	// Verify the mock response contains expected content
	if response.Text != "Hello! How can I help you today?" {
		t.Errorf("Unexpected response content: %s", response.Text)
	}
}