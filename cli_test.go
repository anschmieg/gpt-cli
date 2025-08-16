package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected *CLIConfig
		wantErr  bool
	}{
		{
			name: "default values",
			args: []string{"test prompt"},
			expected: &CLIConfig{
				Provider:    "copilot",
				Temperature: 0.6,
				Verbose:     false,
				Markdown:    true,
				RetryModel:  false,
				Stream:      false,
				Suggest:     false,
				Help:        false,
				Prompt:      "test prompt",
			},
		},
		{
			name: "with provider and model",
			args: []string{"--provider", "openai", "--model", "gpt-4", "hello world"},
			expected: &CLIConfig{
				Provider:    "openai",
				Model:       "gpt-4",
				Temperature: 0.6,
				Verbose:     false,
				Markdown:    true,
				RetryModel:  false,
				Stream:      false,
				Suggest:     false,
				Help:        false,
				Prompt:      "hello world",
			},
		},
		{
			name: "with temperature and verbose",
			args: []string{"--temperature", "0.8", "--verbose", "test"},
			expected: &CLIConfig{
				Provider:    "copilot",
				Temperature: 0.8,
				Verbose:     true,
				Markdown:    true,
				RetryModel:  false,
				Stream:      false,
				Suggest:     false,
				Help:        false,
				Prompt:      "test",
			},
		},
		{
			name: "help flag",
			args: []string{"--help"},
			expected: &CLIConfig{
				Provider:    "copilot",
				Temperature: 0.6,
				Verbose:     false,
				Markdown:    true,
				RetryModel:  false,
				Stream:      false,
				Suggest:     false,
				Help:        true,
				Prompt:      "",
			},
		},
		{
			name: "suggest flag",
			args: []string{"--suggest", "list files"},
			expected: &CLIConfig{
				Provider:    "copilot",
				Temperature: 0.6,
				Verbose:     false,
				Markdown:    true,
				RetryModel:  false,
				Stream:      false,
				Suggest:     true,
				Help:        false,
				Prompt:      "list files",
			},
		},
		{
			name: "stream flag",
			args: []string{"--stream", "test streaming"},
			expected: &CLIConfig{
				Provider:    "copilot",
				Temperature: 0.6,
				Verbose:     false,
				Markdown:    true,
				RetryModel:  false,
				Stream:      true,
				Suggest:     false,
				Help:        false,
				Prompt:      "test streaming",
			},
		},
		{
			name: "all flags combined",
			args: []string{
				"--provider", "openai",
				"--model", "gpt-4",
				"--temperature", "0.9",
				"--system", "custom system",
				"--file", "/path/to/file",
				"--verbose",
				"--markdown=false",
				"--retry-model",
				"--stream",
				"--suggest",
				"complex prompt",
			},
			expected: &CLIConfig{
				Provider:    "openai",
				Model:       "gpt-4",
				Temperature: 0.9,
				System:      "custom system",
				File:        "/path/to/file",
				Verbose:     true,
				Markdown:    false,
				RetryModel:  true,
				Stream:      true,
				Suggest:     true,
				Help:        false,
				Prompt:      "complex prompt",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("parseArgs() = %+v, expected %+v", got, tt.expected)
			}
		})
	}
}

func TestGetDefaultModel(t *testing.T) {
	tests := []struct {
		provider string
		expected string
	}{
		{"openai", "gpt-4o-mini"},
		{"copilot", "gpt-4o-mini"},
		{"gemini", "gemini-2.0-flash"},
		{"unknown", "gpt-4o-mini"},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			got := getDefaultModel(tt.provider)
			if got != tt.expected {
				t.Errorf("getDefaultModel(%s) = %s, expected %s", tt.provider, got, tt.expected)
			}
		})
	}
}

func TestIsModelNotSupportedError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"model not supported", fmt.Errorf("model_not_supported"), true},
		{"model is not supported", fmt.Errorf("The model is not supported"), true},
		{"requested model not supported", fmt.Errorf("requested model is not supported"), true},
		{"different error", fmt.Errorf("network error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isModelNotSupportedError(tt.err)
			if got != tt.expected {
				t.Errorf("isModelNotSupportedError() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestRunCLI(t *testing.T) {
	// Start mock server for testing
	server := startMockTestServer()
	defer server.Close()

	// Set test environment variables
	t.Setenv("GPT_CLI_TEST", "1")
	t.Setenv("MOCK_SERVER_URL", server.URL)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "basic_prompt",
			args:    []string{"hello world"},
			wantErr: false,
		},
		{
			name:    "help_flag",
			args:    []string{"--help"},
			wantErr: false,
		},
		{
			name:    "empty_prompt_no_help",
			args:    []string{},
			wantErr: false, // Should show help
		},
		{
			name:    "suggest_mode",
			args:    []string{"--suggest", "list files"},
			wantErr: false,
		},
		{
			name:    "stream_mode",
			args:    []string{"--stream", "test streaming"},
			wantErr: false,
		},
		{
			name:    "with_provider",
			args:    []string{"--provider", "openai", "test"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runCLI(tt.args)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("runCLI() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApplyConfigDefaults(t *testing.T) {
	fileConfig := &Config{
		DefaultProvider:    "file-provider",
		DefaultModel:       "file-model",
		DefaultTemperature: 0.8,
		DefaultSystem:      "file-system",
	}

	cliConfig := &CLIConfig{
		Provider:    "copilot", // Should be kept
		Model:       "",        // Should be set from file
		Temperature: 0.6,       // Should be kept
		System:      "",        // Should be set from file
	}

	ApplyConfigDefaults(cliConfig, fileConfig)

	// Provider should be overridden by file config since it's non-empty
	if cliConfig.Provider != "file-provider" {
		t.Errorf("Expected provider to be overridden to 'file-provider', got %s", cliConfig.Provider)
	}

	if cliConfig.Model != "file-model" {
		t.Errorf("Expected model to be set to 'file-model', got %s", cliConfig.Model)
	}

	// Temperature should be overridden by file config
	if cliConfig.Temperature != 0.8 {
		t.Errorf("Expected temperature to be overridden to 0.8, got %f", cliConfig.Temperature)
	}

	if cliConfig.System != "file-system" {
		t.Errorf("Expected system to be set to 'file-system', got %s", cliConfig.System)
	}
}