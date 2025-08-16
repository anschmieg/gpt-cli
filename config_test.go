package main

import (
	"testing"
)

func TestStreaming(t *testing.T) {
	// Note: This test would require a more sophisticated mock server
	// that supports SSE streaming. For now, we test that the function
	// exists and can be called.
	
	config := &CoreConfig{
		Provider:    "openai",
		Model:       "gpt-3.5-turbo",
		Temperature: 0.7,
		Prompt:      "test streaming",
		UseMarkdown: true,
		Stream:      true,
	}
	
	opts := &ProviderOptions{
		APIKey:  "test-key",
		BaseURL: "http://invalid-url-for-test",
	}

	// This should fail with a connection error, but not crash
	err := tryStreamingProvider(config, opts)
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}

func TestConfigLoading(t *testing.T) {
	// Test loading config when file doesn't exist
	config, err := LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig() error = %v", err)
	}
	if config == nil {
		t.Error("Expected non-nil config")
	}
}

func TestConfigDefaults(t *testing.T) {
	cliConfig := &CLIConfig{
		Provider:    "copilot",
		Temperature: 0.6,
	}
	
	fileConfig := &Config{
		DefaultProvider:    "openai",
		DefaultTemperature: 0.8,
		DefaultModel:       "gpt-4",
	}
	
	ApplyConfigDefaults(cliConfig, fileConfig)
	
	// Provider should change from default
	if cliConfig.Provider != "openai" {
		t.Errorf("Expected provider to be overridden to 'openai', got %s", cliConfig.Provider)
	}
	
	// Temperature should change from default
	if cliConfig.Temperature != 0.8 {
		t.Errorf("Expected temperature to be overridden to 0.8, got %f", cliConfig.Temperature)
	}
	
	// Model should be set
	if cliConfig.Model != "gpt-4" {
		t.Errorf("Expected model to be set to 'gpt-4', got %s", cliConfig.Model)
	}
}

func TestSaveConfig(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	
	// Override home directory for testing
	t.Setenv("HOME", tmpDir)
	
	config := &Config{
		DefaultProvider: "test-provider",
		DefaultModel:    "test-model",
	}
	
	err := SaveConfig(config)
	if err != nil {
		t.Errorf("SaveConfig failed: %v", err)
	}
	
	// Try to load the saved config
	loadedConfig, err := LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig failed after save: %v", err)
	}
	
	if loadedConfig.DefaultProvider != config.DefaultProvider {
		t.Errorf("Loaded provider doesn't match saved provider")
	}
	
	if loadedConfig.DefaultModel != config.DefaultModel {
		t.Errorf("Loaded model doesn't match saved model")
	}
}