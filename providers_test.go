package main

import (
	"testing"
)

func TestCallOpenAIProvider(t *testing.T) {
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
		System:      "test system",
		Temperature: 0.7,
	}

	opts := &ProviderOptions{
		APIKey: "test-key",
	}

	response, err := callOpenAIProvider(config, opts)
	if err != nil {
		t.Errorf("callOpenAIProvider failed: %v", err)
		return
	}

	if response.Text == "" && response.Markdown == "" {
		t.Error("Expected non-empty response")
	}
}

func TestCallCopilotProvider(t *testing.T) {
	// Start mock server for testing
	server := startMockTestServer()
	defer server.Close()

	// Set test environment variables
	t.Setenv("GPT_CLI_TEST", "1")
	t.Setenv("MOCK_SERVER_URL", server.URL)

	config := &CoreConfig{
		Provider:    "copilot",
		Model:       "gpt-4",
		Prompt:      "test prompt",
		System:      "test system",
		Temperature: 0.7,
	}

	opts := &ProviderOptions{
		APIKey: "test-key",
	}

	response, err := callCopilotProvider(config, opts)
	if err != nil {
		t.Errorf("callCopilotProvider failed: %v", err)
		return
	}

	if response.Text == "" && response.Markdown == "" {
		t.Error("Expected non-empty response")
	}
}

func TestCallGeminiProvider(t *testing.T) {
	// Start mock server for testing
	server := startMockTestServer()
	defer server.Close()

	// Set test environment variables
	t.Setenv("GPT_CLI_TEST", "1")
	t.Setenv("MOCK_SERVER_URL", server.URL)

	config := &CoreConfig{
		Provider:    "gemini",
		Model:       "gemini-2.0-flash",
		Prompt:      "test prompt",
		System:      "test system",
		Temperature: 0.7,
	}

	opts := &ProviderOptions{
		APIKey: "test-key",
	}

	response, err := callGeminiProvider(config, opts)
	if err != nil {
		t.Errorf("callGeminiProvider failed: %v", err)
		return
	}

	if response.Text == "" && response.Markdown == "" {
		t.Error("Expected non-empty response")
	}
}

func TestCallProvider(t *testing.T) {
	// Start mock server for testing
	server := startMockTestServer()
	defer server.Close()

	// Set test environment variables
	t.Setenv("GPT_CLI_TEST", "1")
	t.Setenv("MOCK_SERVER_URL", server.URL)

	tests := []struct {
		name     string
		provider string
		wantErr  bool
	}{
		{
			name:     "openai_provider",
			provider: "openai",
			wantErr:  false,
		},
		{
			name:     "copilot_provider",
			provider: "copilot",
			wantErr:  false,
		},
		{
			name:     "gemini_provider",
			provider: "gemini",
			wantErr:  false,
		},
		{
			name:     "unsupported_provider",
			provider: "unsupported",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &CoreConfig{
				Provider:    tt.provider,
				Model:       "gpt-4",
				Prompt:      "test prompt",
				Temperature: 0.7,
			}

			opts := &ProviderOptions{
				APIKey: "test-key",
			}

			_, err := callProvider(config, opts)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("callProvider() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetMockServerURL(t *testing.T) {
	tests := []struct {
		name           string
		testEnv        string
		mockServerURL  string
		expectedResult string
	}{
		{
			name:           "not_in_test_mode",
			testEnv:        "",
			mockServerURL:  "",
			expectedResult: "",
		},
		{
			name:           "test_mode_with_custom_url",
			testEnv:        "1",
			mockServerURL:  "http://custom:8080",
			expectedResult: "http://custom:8080",
		},
		{
			name:           "test_mode_with_default_url",
			testEnv:        "1",
			mockServerURL:  "",
			expectedResult: "http://127.0.0.1:8086",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.testEnv != "" {
				t.Setenv("GPT_CLI_TEST", tt.testEnv)
			}
			if tt.mockServerURL != "" {
				t.Setenv("MOCK_SERVER_URL", tt.mockServerURL)
			}

			result := getMockServerURL()
			if result != tt.expectedResult {
				t.Errorf("getMockServerURL() = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}