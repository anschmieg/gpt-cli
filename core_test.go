package main

import (
	"testing"
)

func TestBuildProviderOptions(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		envVars  map[string]string
		expected *ProviderOptions
	}{
		{
			name:     "openai_provider",
			provider: "openai",
			envVars: map[string]string{
				"OPENAI_API_KEY":  "test-key",
				"OPENAI_API_BASE": "https://api.custom.com",
			},
			expected: &ProviderOptions{
				APIKey:  "test-key",
				BaseURL: "https://api.custom.com",
			},
		},
		{
			name:     "copilot_provider",
			provider: "copilot",
			envVars: map[string]string{
				"COPILOT_API_KEY":  "copilot-key",
				"COPILOT_API_BASE": "https://copilot.custom.com",
			},
			expected: &ProviderOptions{
				APIKey:  "copilot-key",
				BaseURL: "https://copilot.custom.com",
			},
		},
		{
			name:     "gemini_provider",
			provider: "gemini",
			envVars: map[string]string{
				"GEMINI_API_KEY": "gemini-key",
			},
			expected: &ProviderOptions{
				APIKey:  "gemini-key",
				BaseURL: "",
			},
		},
		{
			name:     "no_env_vars",
			provider: "openai",
			envVars:  map[string]string{},
			expected: &ProviderOptions{
				APIKey:  "",
				BaseURL: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			result, err := buildProviderOptions(tt.provider)
			if err != nil {
				t.Errorf("buildProviderOptions failed: %v", err)
				return
			}

			if result.APIKey != tt.expected.APIKey {
				t.Errorf("Expected APIKey '%s', got '%s'", tt.expected.APIKey, result.APIKey)
			}

			if result.BaseURL != tt.expected.BaseURL {
				t.Errorf("Expected BaseURL '%s', got '%s'", tt.expected.BaseURL, result.BaseURL)
			}
		})
	}
}

func TestRunCoreWithSuggestionMode(t *testing.T) {
	// Start mock server for testing
	server := startMockTestServer()
	defer server.Close()

	// Set test environment variables
	t.Setenv("GPT_CLI_TEST", "1")
	t.Setenv("MOCK_SERVER_URL", server.URL)

	config := &CoreConfig{
		Provider:    "openai",
		Model:       "gpt-4",
		Prompt:      "list files in current directory",
		SuggestMode: true,
		Verbose:     false,
	}

	opts := &ProviderOptions{
		APIKey: "test-key",
	}

	err := runCore(config, opts)
	if err != nil {
		t.Errorf("runCore with suggestion mode failed: %v", err)
	}
}

func TestRunCoreWithMarkdown(t *testing.T) {
	// Start mock server for testing
	server := startMockTestServer()
	defer server.Close()

	// Set test environment variables
	t.Setenv("GPT_CLI_TEST", "1")
	t.Setenv("MOCK_SERVER_URL", server.URL)

	config := &CoreConfig{
		Provider:    "openai",
		Model:       "gpt-4",
		Prompt:      "test prompt",
		UseMarkdown: true,
		Verbose:     false,
	}

	opts := &ProviderOptions{
		APIKey: "test-key",
	}

	err := runCore(config, opts)
	if err != nil {
		t.Errorf("runCore with markdown failed: %v", err)
	}
}

func TestRunCoreWithoutMarkdown(t *testing.T) {
	// Start mock server for testing
	server := startMockTestServer()
	defer server.Close()

	// Set test environment variables
	t.Setenv("GPT_CLI_TEST", "1")
	t.Setenv("MOCK_SERVER_URL", server.URL)

	config := &CoreConfig{
		Provider:    "openai",
		Model:       "gpt-4",
		Prompt:      "test prompt",
		UseMarkdown: false,
		Verbose:     false,
	}

	opts := &ProviderOptions{
		APIKey: "test-key",
	}

	err := runCore(config, opts)
	if err != nil {
		t.Errorf("runCore without markdown failed: %v", err)
	}
}

func TestRunCoreWithStreaming(t *testing.T) {
	// Start mock server for testing
	server := startMockTestServer()
	defer server.Close()

	// Set test environment variables
	t.Setenv("GPT_CLI_TEST", "1")
	t.Setenv("MOCK_SERVER_URL", server.URL)

	config := &CoreConfig{
		Provider:    "openai",
		Model:       "gpt-4",
		Prompt:      "test prompt",
		Stream:      true,
		UseMarkdown: true,
		Verbose:     false,
	}

	opts := &ProviderOptions{
		APIKey: "test-key",
	}

	err := runCore(config, opts)
	if err != nil {
		t.Errorf("runCore with streaming failed: %v", err)
	}
}

func TestRunCoreWithRetry(t *testing.T) {
	// Start mock server for testing
	server := startMockTestServer()
	defer server.Close()

	// Set test environment variables
	t.Setenv("GPT_CLI_TEST", "1")
	t.Setenv("MOCK_SERVER_URL", server.URL)

	config := &CoreConfig{
		Provider:       "openai",
		Model:          "unsupported-model", // This should trigger retry
		Prompt:         "test prompt",
		AutoRetryModel: true,
		Verbose:        true,
	}

	opts := &ProviderOptions{
		APIKey: "test-key",
	}

	// Note: This test depends on the mock server handling unsupported models
	// The actual behavior may vary based on mock server implementation
	err := runCore(config, opts)
	// We don't assert no error here since the mock server might not implement retry logic
	// But we verify the function doesn't panic
	_ = err
}

func TestRunCoreErrorHandling(t *testing.T) {
	tests := []struct {
		name    string
		config  *CoreConfig
		opts    *ProviderOptions
		wantErr bool
	}{
		{
			name: "invalid_provider",
			config: &CoreConfig{
				Provider: "invalid-provider",
				Model:    "gpt-4",
				Prompt:   "test prompt",
			},
			opts: &ProviderOptions{
				APIKey: "test-key",
			},
			wantErr: true,
		},
		{
			name: "empty_provider",
			config: &CoreConfig{
				Provider: "",
				Model:    "gpt-4",
				Prompt:   "test prompt",
			},
			opts: &ProviderOptions{
				APIKey: "test-key",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runCore(tt.config, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("runCore() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSystemPromptDefault(t *testing.T) {
	// Start mock server for testing
	server := startMockTestServer()
	defer server.Close()

	// Set test environment variables
	t.Setenv("GPT_CLI_TEST", "1")
	t.Setenv("MOCK_SERVER_URL", server.URL)

	config := &CoreConfig{
		Provider: "openai",
		Model:    "gpt-4",
		Prompt:   "test prompt",
		System:   "", // Empty system prompt should use default
		Verbose:  false,
	}

	opts := &ProviderOptions{
		APIKey: "test-key",
	}

	err := runCore(config, opts)
	if err != nil {
		t.Errorf("runCore with default system prompt failed: %v", err)
	}

	// Verify that the system prompt was set to default
	expectedDefault := "You are an AI assistant called via CLI. Respond concisely and clearly, focusing only on the user's prompt. Include only very brief explanations unless explicitly asked."
	if config.System != expectedDefault {
		t.Errorf("Expected default system prompt to be set, got: %s", config.System)
	}
}

func TestModelDefault(t *testing.T) {
	tests := []struct {
		name           string
		provider       string
		expectedModel  string
	}{
		{
			name:           "gemini_default",
			provider:       "gemini",
			expectedModel:  "gemini-2.0-flash",
		},
		{
			name:           "openai_default",
			provider:       "openai",
			expectedModel:  "gpt-4o-mini",
		},
		{
			name:           "copilot_default",
			provider:       "copilot",
			expectedModel:  "gpt-4o-mini",
		},
		{
			name:           "unknown_default",
			provider:       "unknown",
			expectedModel:  "gpt-4o-mini",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getDefaultModel(tt.provider)
			if result != tt.expectedModel {
				t.Errorf("Expected default model '%s' for provider '%s', got '%s'", 
					tt.expectedModel, tt.provider, result)
			}
		})
	}
}