package main

import (
	"testing"
)

func TestGetProviderBaseURL(t *testing.T) {
	tests := []struct {
		name        string
		provider    string
		opts        *ProviderOptions
		expected    string
		setMockURL  bool
		mockURL     string
	}{
		{
			name:     "openai_default",
			provider: "openai",
			opts:     &ProviderOptions{},
			expected: "https://api.openai.com",
		},
		{
			name:     "copilot_default",
			provider: "copilot",
			opts:     &ProviderOptions{},
			expected: "https://api.github.com",
		},
		{
			name:     "gemini_default",
			provider: "gemini",
			opts:     &ProviderOptions{},
			expected: "https://generativelanguage.googleapis.com",
		},
		{
			name:     "unknown_provider_default",
			provider: "unknown",
			opts:     &ProviderOptions{},
			expected: "https://api.openai.com",
		},
		{
			name:     "custom_base_url",
			provider: "openai",
			opts: &ProviderOptions{
				BaseURL: "https://custom.api.com",
			},
			expected: "https://custom.api.com",
		},
		{
			name:       "mock_server_url",
			provider:   "openai",
			opts:       &ProviderOptions{},
			setMockURL: true,
			mockURL:    "http://localhost:8080",
			expected:   "http://localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setMockURL {
				t.Setenv("GPT_CLI_TEST", "1")
				t.Setenv("MOCK_SERVER_URL", tt.mockURL)
			}

			result := getProviderBaseURL(tt.provider, tt.opts)
			if result != tt.expected {
				t.Errorf("Expected base URL '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestTryStreamingProvider(t *testing.T) {
	// Start mock server for testing
	server := startMockTestServer()
	defer server.Close()

	// Set test environment variables
	t.Setenv("GPT_CLI_TEST", "1")
	t.Setenv("MOCK_SERVER_URL", server.URL)

	tests := []struct {
		name        string
		config      *CoreConfig
		expectError bool
	}{
		{
			name: "valid_streaming_request",
			config: &CoreConfig{
				Provider:    "openai",
				Model:       "gpt-4",
				Prompt:      "test prompt",
				System:      "test system",
				Temperature: 0.7,
				UseMarkdown: false,
				Verbose:     false,
			},
			expectError: false,
		},
		{
			name: "streaming_with_markdown",
			config: &CoreConfig{
				Provider:    "openai",
				Model:       "gpt-4",
				Prompt:      "test prompt with **markdown**",
				System:      "test system",
				Temperature: 0.7,
				UseMarkdown: true,
				Verbose:     false,
			},
			expectError: false,
		},
		{
			name: "empty_prompt",
			config: &CoreConfig{
				Provider:    "openai",
				Model:       "gpt-4",
				Prompt:      "",
				System:      "test system",
				Temperature: 0.7,
				UseMarkdown: false,
				Verbose:     false,
			},
			expectError: false, // Should not error, might just return empty response
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &ProviderOptions{
				APIKey: "test-key",
			}

			err := tryStreamingProvider(tt.config, opts)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestStreamChatCompletion(t *testing.T) {
	// Start mock server for testing
	server := startMockTestServer()
	defer server.Close()

	// Set test environment variables
	t.Setenv("GPT_CLI_TEST", "1")
	t.Setenv("MOCK_SERVER_URL", server.URL)

	tests := []struct {
		name     string
		config   *CoreConfig
		opts     *ProviderOptions
		wantErr  bool
	}{
		{
			name: "successful_stream",
			config: &CoreConfig{
				Provider:    "openai",
				Model:       "gpt-4",
				Prompt:      "Hello",
				System:      "You are a helpful assistant",
				Temperature: 0.7,
				UseMarkdown: false,
				Verbose:     false,
			},
			opts: &ProviderOptions{
				APIKey: "test-key",
			},
			wantErr: false,
		},
		{
			name: "stream_with_markdown_rendering",
			config: &CoreConfig{
				Provider:    "openai",
				Model:       "gpt-4",
				Prompt:      "Generate some markdown",
				System:      "You are a helpful assistant",
				Temperature: 0.7,
				UseMarkdown: true,
				Verbose:     false,
			},
			opts: &ProviderOptions{
				APIKey: "test-key",
			},
			wantErr: false,
		},
		{
			name: "verbose_logging",
			config: &CoreConfig{
				Provider:    "openai",
				Model:       "gpt-4",
				Prompt:      "Test verbose",
				System:      "You are a helpful assistant",
				Temperature: 0.7,
				UseMarkdown: false,
				Verbose:     true,
			},
			opts: &ProviderOptions{
				APIKey: "test-key",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := streamChatCompletion(tt.config, tt.opts)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("streamChatCompletion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStreamChatCompletionErrorHandling(t *testing.T) {
	tests := []struct {
		name    string
		config  *CoreConfig
		opts    *ProviderOptions
		wantErr bool
	}{
		{
			name: "invalid_url",
			config: &CoreConfig{
				Provider:    "openai",
				Model:       "gpt-4",
				Prompt:      "test",
				Temperature: 0.7,
			},
			opts: &ProviderOptions{
				BaseURL: "invalid-url-for-test", // This should cause an error
				APIKey:  "test-key",
			},
			wantErr: true,
		},
		{
			name: "missing_api_key",
			config: &CoreConfig{
				Provider:    "openai",
				Model:       "gpt-4",
				Prompt:      "test",
				Temperature: 0.7,
			},
			opts: &ProviderOptions{
				BaseURL: "https://api.openai.com",
				APIKey:  "", // Missing API key
			},
			wantErr: true, // May or may not error depending on server
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := streamChatCompletion(tt.config, tt.opts)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("streamChatCompletion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}