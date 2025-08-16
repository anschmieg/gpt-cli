package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents the configuration file structure
type Config struct {
	DefaultProvider    string            `json:"default_provider,omitempty"`
	DefaultModel       string            `json:"default_model,omitempty"`
	DefaultTemperature float64           `json:"default_temperature,omitempty"`
	DefaultSystem      string            `json:"default_system,omitempty"`
	ProviderSettings   map[string]string `json:"provider_settings,omitempty"`
}

// LoadConfig loads configuration from ~/.gpt-cli/config.json if it exists
func LoadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return &Config{}, nil // Return empty config if can't get home dir
	}

	configPath := filepath.Join(homeDir, ".gpt-cli", "config.json")
	
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{}, nil // Return empty config if file doesn't exist
	}

	// Read and parse config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return &Config{}, nil // Return empty config if can't read file
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return &Config{}, nil // Return empty config if can't parse JSON
	}

	return &config, nil
}

// SaveConfig saves configuration to ~/.gpt-cli/config.json
func SaveConfig(config *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".gpt-cli")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.json")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// ApplyConfigDefaults applies configuration defaults to CLI config
func ApplyConfigDefaults(cliConfig *CLIConfig, fileConfig *Config) {
	if cliConfig.Provider == "copilot" && fileConfig.DefaultProvider != "" {
		cliConfig.Provider = fileConfig.DefaultProvider
	}
	if cliConfig.Model == "" && fileConfig.DefaultModel != "" {
		cliConfig.Model = fileConfig.DefaultModel
	}
	if cliConfig.Temperature == 0.6 && fileConfig.DefaultTemperature != 0 {
		cliConfig.Temperature = fileConfig.DefaultTemperature
	}
	if cliConfig.System == "" && fileConfig.DefaultSystem != "" {
		cliConfig.System = fileConfig.DefaultSystem
	}
}