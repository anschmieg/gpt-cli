package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected *Config
	}{
		{
			name:    "default values",
			envVars: map[string]string{},
			expected: &Config{
				Provider:    DefaultProvider,
				Model:       DefaultModel,
				Temperature: DefaultTemperature,
				Verbose:     false,
				Markdown:    true,
				System:      DefaultSystem,
			},
		},
		{
			name: "environment overrides",
			envVars: map[string]string{
				"GPT_CLI_PROVIDER":    "openai",
				"GPT_CLI_MODEL":       "gpt-4",
				"GPT_CLI_TEMPERATURE": "0.8",
				"GPT_CLI_VERBOSE":     "true",
				"GPT_CLI_MARKDOWN":    "false",
				"OPENAI_API_KEY":      "test-key",
			},
			expected: &Config{
				Provider:    "openai",
				Model:       "gpt-4",
				Temperature: 0.8,
				Verbose:     true,
				Markdown:    false,
				System:      DefaultSystem,
				APIKey:      "test-key",
				BaseURL:     "https://api.openai.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			clearTestEnv()

			// Set test environment
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			config := NewConfig()

			assert.Equal(t, tt.expected.Provider, config.Provider)
			assert.Equal(t, tt.expected.Model, config.Model)
			assert.Equal(t, tt.expected.Temperature, config.Temperature)
			assert.Equal(t, tt.expected.Verbose, config.Verbose)
			assert.Equal(t, tt.expected.Markdown, config.Markdown)
			assert.Equal(t, tt.expected.System, config.System)
			assert.Equal(t, tt.expected.APIKey, config.APIKey)
			assert.Equal(t, tt.expected.BaseURL, config.BaseURL)

			// Cleanup
			clearTestEnv()
		})
	}
}

func TestLoadConfigFromPath(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		filename    string
		expected    *FileConfig
		expectError bool
	}{
		{
			name:     "valid YAML config",
			filename: "config.yml",
			content: `
provider: openai
model: gpt-4
temperature: 0.7
verbose: true
markdown: false
system: "Test system"
providers:
  openai:
    api_key: test-key
    base_url: https://api.openai.com/v1
`,
			expected: &FileConfig{
				Provider:    "openai",
				Model:       "gpt-4",
				Temperature: 0.7,
				Verbose:     true,
				Markdown:    false,
				System:      "Test system",
				Providers: ProviderConfig{
					OpenAI: struct {
						APIKey  string `json:"api_key" yaml:"api_key"`
						BaseURL string `json:"base_url" yaml:"base_url"`
					}{
						APIKey:  "test-key",
						BaseURL: "https://api.openai.com/v1",
					},
				},
			},
		},
		{
			name:     "valid JSON config",
			filename: "config.json",
			content: `{
  "provider": "copilot",
  "model": "gpt-4o-mini",
  "temperature": 0.5,
  "verbose": false,
  "markdown": true,
  "system": "JSON system",
  "providers": {
    "copilot": {
      "api_key": "copilot-key",
      "base_url": "https://api.copilot.com"
    }
  }
}`,
			expected: &FileConfig{
				Provider:    "copilot",
				Model:       "gpt-4o-mini",
				Temperature: 0.5,
				Verbose:     false,
				Markdown:    true,
				System:      "JSON system",
				Providers: ProviderConfig{
					Copilot: struct {
						APIKey  string `json:"api_key" yaml:"api_key"`
						BaseURL string `json:"base_url" yaml:"base_url"`
					}{
						APIKey:  "copilot-key",
						BaseURL: "https://api.copilot.com",
					},
				},
			},
		},
		{
			name:        "invalid YAML",
			filename:    "config.yml",
			content:     "invalid: yaml: content:",
			expectError: true,
		},
		{
			name:        "invalid JSON",
			filename:    "config.json",
			content:     `{"invalid": json}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, tt.filename)

			err := os.WriteFile(configPath, []byte(tt.content), 0644)
			require.NoError(t, err)

			// Test loading
			config, err := loadConfigFromPath(configPath)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, config)
			}
		})
	}
}

func TestLoadConfigFromPathEnvSubstitution(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "env-key")
	os.Setenv("OPENAI_API_BASE", "https://env-base")
	defer os.Unsetenv("OPENAI_API_KEY")
	defer os.Unsetenv("OPENAI_API_BASE")

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")
	content := `
provider: openai
providers:
  openai:
    api_key: ${OPENAI_API_KEY}
    base_url: ${OPENAI_API_BASE}
`
	os.WriteFile(configPath, []byte(content), 0644)

	cfg, err := loadConfigFromPath(configPath)
	require.NoError(t, err)
	assert.Equal(t, "env-key", cfg.Providers.OpenAI.APIKey)
	assert.Equal(t, "https://env-base", cfg.Providers.OpenAI.BaseURL)
}
func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "environment variable set",
			key:          "TEST_VAR",
			defaultValue: "default",
			envValue:     "env_value",
			expected:     "env_value",
		},
		{
			name:         "environment variable not set",
			key:          "TEST_VAR",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnvOrDefault(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvFloatOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue float64
		envValue     string
		expected     float64
	}{
		{
			name:         "valid float",
			key:          "TEST_FLOAT",
			defaultValue: 1.0,
			envValue:     "2.5",
			expected:     2.5,
		},
		{
			name:         "invalid float",
			key:          "TEST_FLOAT",
			defaultValue: 1.0,
			envValue:     "invalid",
			expected:     1.0,
		},
		{
			name:         "not set",
			key:          "TEST_FLOAT",
			defaultValue: 1.0,
			envValue:     "",
			expected:     1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnvFloatOrDefault(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvBoolOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue bool
		envValue     string
		expected     bool
	}{
		{
			name:         "true value",
			key:          "TEST_BOOL",
			defaultValue: false,
			envValue:     "true",
			expected:     true,
		},
		{
			name:         "false value",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "false",
			expected:     false,
		},
		{
			name:         "invalid value",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "invalid",
			expected:     true,
		},
		{
			name:         "not set",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "",
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnvBoolOrDefault(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetConfigDir(t *testing.T) {
	tests := []struct {
		name     string
		xdgHome  string
		home     string
		expected string
	}{
		{
			name:     "XDG_CONFIG_HOME set",
			xdgHome:  "/custom/config",
			home:     "/home/user",
			expected: "/custom/config",
		},
		{
			name:     "HOME set, no XDG",
			xdgHome:  "",
			home:     "/home/user",
			expected: "/home/user/.config",
		},
		{
			name:     "neither set",
			xdgHome:  "",
			home:     "",
			expected: ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			origXDG := os.Getenv("XDG_CONFIG_HOME")
			origHome := os.Getenv("HOME")

			// Set test values
			if tt.xdgHome != "" {
				os.Setenv("XDG_CONFIG_HOME", tt.xdgHome)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}

			if tt.home != "" {
				os.Setenv("HOME", tt.home)
			} else {
				os.Unsetenv("HOME")
			}

			result := getConfigDir()
			assert.Equal(t, tt.expected, result)

			// Restore original values
			if origXDG != "" {
				os.Setenv("XDG_CONFIG_HOME", origXDG)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}

			if origHome != "" {
				os.Setenv("HOME", origHome)
			} else {
				os.Unsetenv("HOME")
			}
		})
	}
}

// clearTestEnv clears test environment variables
func clearTestEnv() {
	testEnvVars := []string{
		"GPT_CLI_PROVIDER",
		"GPT_CLI_MODEL",
		"GPT_CLI_TEMPERATURE",
		"GPT_CLI_VERBOSE",
		"GPT_CLI_MARKDOWN",
		"GPT_CLI_SYSTEM",
		"OPENAI_API_KEY",
		"OPENAI_API_BASE",
		"COPILOT_API_KEY",
		"COPILOT_API_BASE",
		"GEMINI_API_KEY",
		"GEMINI_API_BASE",
	}

	for _, envVar := range testEnvVars {
		os.Unsetenv(envVar)
	}
}
