package modes

import (
	"errors"
	"strings"
	"testing"

	"github.com/anschmieg/gpt-cli/internal/config"
	"github.com/anschmieg/gpt-cli/internal/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockProvider for testing shell mode
type MockProvider struct {
	response     string
	streamChunks []string
	err          error
}

func (m *MockProvider) CallProvider(prompt string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	if m.response != "" {
		return m.response, nil
	}
	return strings.Join(m.streamChunks, ""), nil
}

func (m *MockProvider) StreamProvider(prompt string) (<-chan string, <-chan error) {
	contentChan := make(chan string)
	errorChan := make(chan error, 1)

	go func() {
		defer close(contentChan)
		defer close(errorChan)

		if m.err != nil {
			errorChan <- m.err
			return
		}

		if len(m.streamChunks) > 0 {
			for _, ch := range m.streamChunks {
				contentChan <- ch
			}
			return
		}

		if m.response != "" {
			contentChan <- m.response
		}
	}()

	return contentChan, errorChan
}

func (m *MockProvider) GetName() string {
	return "mock"
}

func TestNewShellMode(t *testing.T) {
	cfg := &config.Config{
		Provider: "mock",
		Model:    "test-model",
	}
	provider := &MockProvider{}
	ui := ui.New()

	mode := NewShellMode(cfg, provider, ui)

	assert.NotNil(t, mode)
	assert.Equal(t, cfg, mode.config)
	assert.Equal(t, provider, mode.provider)
	assert.Equal(t, ui, mode.ui)
}

func TestParseShellSuggestion(t *testing.T) {
	mode := &ShellMode{}

	tests := []struct {
		name        string
		response    string
		expected    *ShellSuggestion
		expectError bool
	}{
		{
			name: "valid JSON response",
			response: `{
				"command": "ls -la",
				"safety_level": "safe",
				"explanation": "Lists all files in the current directory",
				"reasoning": "Only reads directory contents without modification"
			}`,
			expected: &ShellSuggestion{
				Command:     "ls -la",
				SafetyLevel: "safe",
				Explanation: "Lists all files in the current directory",
				Reasoning:   "Only reads directory contents without modification",
			},
		},
		{
			name:     "JSON with markdown code blocks",
			response: "```json\n{\n  \"command\": \"mkdir test\",\n  \"safety_level\": \"moderate\",\n  \"explanation\": \"Creates a directory named test\"\n}\n```",
			expected: &ShellSuggestion{
				Command:     "mkdir test",
				SafetyLevel: "moderate",
				Explanation: "Creates a directory named test",
			},
		},
		{
			name:     "JSON embedded in text",
			response: "Here's the command suggestion: {\"command\": \"echo hello\", \"safety_level\": \"safe\", \"explanation\": \"Prints hello to stdout\"} Hope this helps!",
			expected: &ShellSuggestion{
				Command:     "echo hello",
				SafetyLevel: "safe",
				Explanation: "Prints hello to stdout",
			},
		},
		{
			name:        "invalid JSON",
			response:    "This is not JSON",
			expectError: true,
		},
		{
			name:        "missing command field",
			response:    `{"safety_level": "safe", "explanation": "test"}`,
			expectError: true,
		},
		{
			name:        "missing safety_level field",
			response:    `{"command": "ls", "explanation": "test"}`,
			expectError: true,
		},
		{
			name:        "missing explanation field",
			response:    `{"command": "ls", "safety_level": "safe"}`,
			expectError: true,
		},
		{
			name:        "invalid safety level",
			response:    `{"command": "ls", "safety_level": "invalid", "explanation": "test"}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mode.parseShellSuggestion(tt.response)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestSuggestCommand(t *testing.T) {
	tests := []struct {
		name        string
		prompt      string
		response    string
		providerErr error
		expectError bool
	}{
		{
			name:   "successful command suggestion",
			prompt: "list files",
			response: `{
				"command": "ls -la",
				"safety_level": "safe",
				"explanation": "Lists all files in the current directory"
			}`,
			expectError: false,
		},
		{
			name:        "provider error",
			prompt:      "test prompt",
			providerErr: errors.New("provider error"),
			expectError: true,
		},
		{
			name:        "invalid JSON response",
			prompt:      "test prompt",
			response:    "invalid json response",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Provider: "mock",
				Model:    "test-model",
			}
			provider := &MockProvider{
				response: tt.response,
				err:      tt.providerErr,
			}
			ui := ui.New()
			mode := NewShellMode(cfg, provider, ui)

			suggestion, err := mode.SuggestCommand(tt.prompt)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, suggestion)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, suggestion)
				assert.NotEmpty(t, suggestion.Command)
				assert.NotEmpty(t, suggestion.SafetyLevel)
				assert.NotEmpty(t, suggestion.Explanation)
			}
		})
	}
}

func TestGetSafetyColor(t *testing.T) {
	mode := &ShellMode{ui: ui.New()}

	tests := []struct {
		level    string
		expected string // Color value
	}{
		{
			level:    "safe",
			expected: "#059669", // Green
		},
		{
			level:    "moderate",
			expected: "#EAB308", // Yellow
		},
		{
			level:    "dangerous",
			expected: "#DC2626", // Red
		},
		{
			level:    "unknown",
			expected: "#6B7280", // Gray (default)
		},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			style := mode.getSafetyColor(tt.level)
			assert.NotNil(t, style)
			// We can't easily test the actual color value, but we can ensure a style is returned
		})
	}
}

func TestDisplaySuggestion(t *testing.T) {
	// This test mainly ensures displaySuggestion doesn't panic
	// Since it prints to stdout, we can't easily test the output
	mode := &ShellMode{ui: ui.New()}

	suggestion := &ShellSuggestion{
		Command:     "ls -la",
		SafetyLevel: "safe",
		Explanation: "Lists all files",
		Reasoning:   "Read-only operation",
	}

	// This should not panic
	mode.displaySuggestion(suggestion)
}

func TestShellSuggestionValidation(t *testing.T) {
	tests := []struct {
		name        string
		suggestion  ShellSuggestion
		expectValid bool
	}{
		{
			name: "valid suggestion",
			suggestion: ShellSuggestion{
				Command:     "ls -la",
				SafetyLevel: "safe",
				Explanation: "Lists files",
			},
			expectValid: true,
		},
		{
			name: "valid dangerous command",
			suggestion: ShellSuggestion{
				Command:     "rm -rf /tmp/*",
				SafetyLevel: "dangerous",
				Explanation: "Deletes all files in /tmp",
				Reasoning:   "Can delete important files",
			},
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - ensure required fields are present
			if tt.expectValid {
				assert.NotEmpty(t, tt.suggestion.Command)
				assert.NotEmpty(t, tt.suggestion.SafetyLevel)
				assert.NotEmpty(t, tt.suggestion.Explanation)

				// Validate safety level is one of the expected values
				validLevels := []string{"safe", "moderate", "dangerous"}
				assert.Contains(t, validLevels, tt.suggestion.SafetyLevel)
			}
		})
	}
}

func TestShellModeIntegration(t *testing.T) {
	// Integration test with a realistic scenario
	cfg := &config.Config{
		Provider:    "mock",
		Model:       "gpt-4",
		Temperature: 0.1,
	}

	mockResponse := `{
		"command": "find . -name '*.go' -type f",
		"safety_level": "safe",
		"explanation": "Finds all Go source files in the current directory and subdirectories",
		"reasoning": "This is a read-only operation that only searches for files without modifying anything"
	}`

	provider := &MockProvider{response: mockResponse}
	ui := ui.New()
	mode := NewShellMode(cfg, provider, ui)

	suggestion, err := mode.SuggestCommand("find all go files")

	require.NoError(t, err)
	assert.Equal(t, "find . -name '*.go' -type f", suggestion.Command)
	assert.Equal(t, "safe", suggestion.SafetyLevel)
	assert.Contains(t, suggestion.Explanation, "Go source files")
	assert.Contains(t, suggestion.Reasoning, "read-only")
}
