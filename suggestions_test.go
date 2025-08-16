package main

import (
	"strings"
	"testing"
)

func TestGenerateSuggestionPrompt(t *testing.T) {
	prompt := generateSuggestionPrompt()
	
	// Check that the prompt contains key elements
	if !strings.Contains(prompt, "shell command assistant") {
		t.Error("Prompt should mention shell command assistant")
	}
	
	if !strings.Contains(prompt, "JSON") {
		t.Error("Prompt should mention JSON format")
	}
	
	if !strings.Contains(prompt, "risk") {
		t.Error("Prompt should mention risk levels")
	}
	
	if !strings.Contains(prompt, "suggestions") {
		t.Error("Prompt should mention suggestions structure")
	}
}

func TestParseSuggestionResponse(t *testing.T) {
	tests := []struct {
		name        string
		response    string
		expectError bool
		expectCount int
	}{
		{
			name: "valid_response",
			response: `{
				"suggestions": [
					{
						"command": "ls -la",
						"description": "List all files",
						"category": "file_management",
						"risk": "low",
						"args": ["-la"]
					}
				],
				"context": "Listing files",
				"safe": true
			}`,
			expectError: false,
			expectCount: 1,
		},
		{
			name: "response_with_markdown",
			response: "```json\n" + `{
				"suggestions": [
					{
						"command": "pwd",
						"description": "Show current directory",
						"risk": "low"
					}
				],
				"safe": true
			}` + "\n```",
			expectError: false,
			expectCount: 1,
		},
		{
			name: "empty_suggestions",
			response: `{
				"suggestions": [],
				"safe": true
			}`,
			expectError: true,
			expectCount: 0,
		},
		{
			name: "invalid_json",
			response: "not valid json",
			expectError: true,
			expectCount: 0,
		},
		{
			name: "missing_command",
			response: `{
				"suggestions": [
					{
						"description": "List files",
						"risk": "low"
					}
				],
				"safe": true
			}`,
			expectError: true,
			expectCount: 0,
		},
		{
			name: "missing_description",
			response: `{
				"suggestions": [
					{
						"command": "ls",
						"risk": "low"
					}
				],
				"safe": true
			}`,
			expectError: true,
			expectCount: 0,
		},
		{
			name: "default_risk_level",
			response: `{
				"suggestions": [
					{
						"command": "ls",
						"description": "List files"
					}
				],
				"safe": true
			}`,
			expectError: false,
			expectCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseSuggestionResponse(tt.response)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if len(result.Suggestions) != tt.expectCount {
				t.Errorf("Expected %d suggestions, got %d", tt.expectCount, len(result.Suggestions))
			}
			
			// Verify default risk level is set
			if tt.name == "default_risk_level" && result.Suggestions[0].Risk != "medium" {
				t.Errorf("Expected default risk level 'medium', got '%s'", result.Suggestions[0].Risk)
			}
		})
	}
}

func TestFormatSuggestionOutput(t *testing.T) {
	suggestions := &SuggestionResponse{
		Suggestions: []Suggestion{
			{
				Command:     "ls -la",
				Description: "List all files",
				Category:    "file_management",
				Risk:        "low",
				Args:        []string{"-la"},
			},
		},
		Context: "Listing files",
		Safe:    true,
	}
	
	output := formatSuggestionOutput(suggestions)
	
	// Check that output is valid JSON
	if !strings.Contains(output, "\"command\"") {
		t.Error("Output should contain command field")
	}
	
	if !strings.Contains(output, "ls -la") {
		t.Error("Output should contain the command")
	}
	
	if !strings.Contains(output, "List all files") {
		t.Error("Output should contain the description")
	}
	
	// Check JSON formatting (should be indented)
	if !strings.Contains(output, "  ") {
		t.Error("Output should be indented JSON")
	}
}

func TestRunSuggestionMode(t *testing.T) {
	// Start mock server for testing
	server := startMockTestServer()
	defer server.Close()
	
	// Set test environment variables
	t.Setenv("GPT_CLI_TEST", "1")
	t.Setenv("MOCK_SERVER_URL", server.URL)
	
	config := &CoreConfig{
		Provider:    "openai",
		Model:       "gpt-4",
		Prompt:      "list files",
		SuggestMode: true,
		Verbose:     false,
	}
	
	opts := &ProviderOptions{
		APIKey: "test-key",
	}
	
	// This should not return an error
	err := runSuggestionMode(config, opts)
	if err != nil {
		t.Errorf("runSuggestionMode failed: %v", err)
	}
}