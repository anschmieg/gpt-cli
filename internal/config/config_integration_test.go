//go:build integration
// +build integration

package config

import (
    "io/fs"
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

func TestConfig_LoadsFromConfigFile_Integration(t *testing.T) {
    dir, err := os.MkdirTemp("", "gptcli_cfg_*")
    if err != nil { t.Fatal(err) }
    defer os.RemoveAll(dir)

    // Point config dir to temp and write YAML
    os.Setenv("XDG_CONFIG_HOME", dir)
    defer os.Unsetenv("XDG_CONFIG_HOME")
    cfgDir := dir + "/gpt-cli"
    if err := os.MkdirAll(cfgDir, fs.ModePerm); err != nil { t.Fatal(err) }
    yaml := []byte("provider: openai\nmodel: gpt-4\ntemperature: 0.5\nproviders:\n  openai:\n    api_key: from-file\n")
    if err := os.WriteFile(cfgDir+"/config.yaml", yaml, 0o644); err != nil { t.Fatal(err) }

    // Ensure env doesn’t override file for BaseURL
    os.Unsetenv("OPENAI_API_BASE")
    os.Setenv("OPENAI_API_KEY", "from-env")
    defer os.Unsetenv("OPENAI_API_KEY")

    cfg := NewConfig()
    assert.Equal(t, "openai", cfg.Provider)
    assert.Equal(t, "gpt-4", cfg.Model)
    assert.InDelta(t, 0.5, cfg.Temperature, 1e-9)
    // env api key overrides file
    assert.Equal(t, "from-env", cfg.APIKey)
    // base URL defaults filled
    assert.Equal(t, "https://api.openai.com", cfg.BaseURL)
}
