//go:build integration
// +build integration

package config

import (
    "os"
    "testing"
    "github.com/stretchr/testify/assert"
)

// Verify env var merging and defaults together (module-level interaction)
func TestConfig_EnvMerging_Integration(t *testing.T) {
    os.Setenv("GPT_CLI_PROVIDER", "openai")
    os.Setenv("GPT_CLI_MODEL", "gpt-4o-mini")
    os.Setenv("GPT_CLI_TEMPERATURE", "0.7")
    defer func(){ os.Unsetenv("GPT_CLI_PROVIDER"); os.Unsetenv("GPT_CLI_MODEL"); os.Unsetenv("GPT_CLI_TEMPERATURE") }()

    cfg := NewConfig()
    assert.Equal(t, "openai", cfg.Provider)
    assert.Equal(t, "gpt-4o-mini", cfg.Model)
    assert.InDelta(t, 0.7, cfg.Temperature, 1e-9)
}

