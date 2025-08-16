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
				Help:        true,
				Prompt:      "",
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