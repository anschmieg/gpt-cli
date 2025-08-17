package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	Provider    string
	Model       string
	Temperature float64
	Verbose     bool
	Markdown    bool
	System      string
	APIKey      string
	BaseURL     string
}

// Default values
const (
	DefaultProvider    = "copilot"
	DefaultModel       = "gpt-4o-mini"
	DefaultTemperature = 0.6
	DefaultSystem      = "You are an AI assistant called via CLI. Respond concisely and clearly, focusing only on the user's prompt. Include only very brief explanations unless explicitly asked."
)

// NewConfig creates a new configuration with defaults
func NewConfig() *Config {
	config := &Config{
		Provider:    getEnvOrDefault("GPT_CLI_PROVIDER", DefaultProvider),
		Model:       getEnvOrDefault("GPT_CLI_MODEL", DefaultModel),
		Temperature: getEnvFloatOrDefault("GPT_CLI_TEMPERATURE", DefaultTemperature),
		Verbose:     getEnvBoolOrDefault("GPT_CLI_VERBOSE", false),
		Markdown:    getEnvBoolOrDefault("GPT_CLI_MARKDOWN", true),
		System:      getEnvOrDefault("GPT_CLI_SYSTEM", DefaultSystem),
	}

	// Set provider-specific configurations
	switch config.Provider {
	case "openai":
		config.APIKey = os.Getenv("OPENAI_API_KEY")
		config.BaseURL = getEnvOrDefault("OPENAI_API_BASE", "https://api.openai.com")
	case "copilot":
		config.APIKey = os.Getenv("COPILOT_API_KEY")
		config.BaseURL = os.Getenv("COPILOT_API_BASE")
	case "gemini":
		config.APIKey = os.Getenv("GEMINI_API_KEY")
		config.BaseURL = getEnvOrDefault("GEMINI_API_BASE", "https://generativelanguage.googleapis.com/v1beta/openai")
	}

	return config
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvFloatOrDefault returns environment variable as float or default
func getEnvFloatOrDefault(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getEnvBoolOrDefault returns environment variable as bool or default
func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}