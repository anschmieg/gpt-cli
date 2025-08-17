package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	Provider    string  `json:"provider" yaml:"provider"`
	Model       string  `json:"model" yaml:"model"`
	Temperature float64 `json:"temperature" yaml:"temperature"`
	Verbose     bool    `json:"verbose" yaml:"verbose"`
	Markdown    bool    `json:"markdown" yaml:"markdown"`
	System      string  `json:"system" yaml:"system"`
	APIKey      string  `json:"api_key" yaml:"api_key"`
	BaseURL     string  `json:"base_url" yaml:"base_url"`
}

// ProviderConfig holds provider-specific configuration
type ProviderConfig struct {
	OpenAI struct {
		APIKey  string `json:"api_key" yaml:"api_key"`
		BaseURL string `json:"base_url" yaml:"base_url"`
	} `json:"openai" yaml:"openai"`
	Copilot struct {
		APIKey  string `json:"api_key" yaml:"api_key"`
		BaseURL string `json:"base_url" yaml:"base_url"`
	} `json:"copilot" yaml:"copilot"`
	Gemini struct {
		APIKey  string `json:"api_key" yaml:"api_key"`
		BaseURL string `json:"base_url" yaml:"base_url"`
	} `json:"gemini" yaml:"gemini"`
}

// FileConfig represents the structure of config files
type FileConfig struct {
	Provider    string         `json:"provider" yaml:"provider"`
	Model       string         `json:"model" yaml:"model"`
	Temperature float64        `json:"temperature" yaml:"temperature"`
	Verbose     bool           `json:"verbose" yaml:"verbose"`
	Markdown    bool           `json:"markdown" yaml:"markdown"`
	System      string         `json:"system" yaml:"system"`
	Providers   ProviderConfig `json:"providers" yaml:"providers"`
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

	// Try to load config from file
	if fileConfig, err := loadConfigFile(); err == nil {
		mergeFileConfig(config, fileConfig)
	}

	// Set provider-specific configurations
	setProviderConfig(config)

	return config
}

// loadConfigFile attempts to load configuration from file
func loadConfigFile() (*FileConfig, error) {
	configDir := getConfigDir()
	
	// Try YAML first, then JSON
	for _, ext := range []string{"yml", "yaml", "json"} {
		configPath := filepath.Join(configDir, "gpt-cli", "config."+ext)
		if config, err := loadConfigFromPath(configPath); err == nil {
			return config, nil
		}
	}
	
	return nil, fmt.Errorf("no config file found")
}

// loadConfigFromPath loads config from a specific file path
func loadConfigFromPath(path string) (*FileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config FileConfig
	
	// Determine format by extension
	ext := filepath.Ext(path)
	switch ext {
	case ".yml", ".yaml":
		err = yaml.Unmarshal(data, &config)
	case ".json":
		err = json.Unmarshal(data, &config)
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", ext)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}
	
	return &config, nil
}

// mergeFileConfig merges file configuration into the main config
func mergeFileConfig(config *Config, fileConfig *FileConfig) {
	if fileConfig.Provider != "" {
		config.Provider = fileConfig.Provider
	}
	if fileConfig.Model != "" {
		config.Model = fileConfig.Model
	}
	if fileConfig.Temperature > 0 {
		config.Temperature = fileConfig.Temperature
	}
	if fileConfig.System != "" {
		config.System = fileConfig.System
	}
	// Note: Only update bool values if they differ from defaults
	if fileConfig.Verbose {
		config.Verbose = fileConfig.Verbose
	}
	if !fileConfig.Markdown {
		config.Markdown = fileConfig.Markdown
	}
}

// setProviderConfig sets provider-specific configuration
func setProviderConfig(config *Config) {
	// Load from config file if it exists
	if fileConfig, err := loadConfigFile(); err == nil {
		switch config.Provider {
		case "openai":
			if fileConfig.Providers.OpenAI.APIKey != "" {
				config.APIKey = fileConfig.Providers.OpenAI.APIKey
			}
			if fileConfig.Providers.OpenAI.BaseURL != "" {
				config.BaseURL = fileConfig.Providers.OpenAI.BaseURL
			}
		case "copilot":
			if fileConfig.Providers.Copilot.APIKey != "" {
				config.APIKey = fileConfig.Providers.Copilot.APIKey
			}
			if fileConfig.Providers.Copilot.BaseURL != "" {
				config.BaseURL = fileConfig.Providers.Copilot.BaseURL
			}
		case "gemini":
			if fileConfig.Providers.Gemini.APIKey != "" {
				config.APIKey = fileConfig.Providers.Gemini.APIKey
			}
			if fileConfig.Providers.Gemini.BaseURL != "" {
				config.BaseURL = fileConfig.Providers.Gemini.BaseURL
			}
		}
	}

	// Environment variables override config file
	switch config.Provider {
	case "openai":
		if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
			config.APIKey = apiKey
		}
		if baseURL := os.Getenv("OPENAI_API_BASE"); baseURL != "" {
			config.BaseURL = baseURL
		} else if config.BaseURL == "" {
			config.BaseURL = "https://api.openai.com"
		}
	case "copilot":
		if apiKey := os.Getenv("COPILOT_API_KEY"); apiKey != "" {
			config.APIKey = apiKey
		}
		if baseURL := os.Getenv("COPILOT_API_BASE"); baseURL != "" {
			config.BaseURL = baseURL
		}
	case "gemini":
		if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
			config.APIKey = apiKey
		}
		if baseURL := os.Getenv("GEMINI_API_BASE"); baseURL != "" {
			config.BaseURL = baseURL
		} else if config.BaseURL == "" {
			config.BaseURL = "https://generativelanguage.googleapis.com/v1beta/openai"
		}
	}
}

// getConfigDir returns the configuration directory
func getConfigDir() string {
	if configDir := os.Getenv("XDG_CONFIG_HOME"); configDir != "" {
		return configDir
	}
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".config")
	}
	return "."
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