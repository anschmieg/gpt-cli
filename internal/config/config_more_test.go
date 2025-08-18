package config

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestMergeFileConfigSelective(t *testing.T) {
    cfg := &Config{Provider: "copilot", Model: "gpt-4o-mini", Temperature: 0.6, Verbose: false, Markdown: true, System: "S"}
    fc := &FileConfig{Provider: "openai", Model: "gpt-4", Temperature: 0.8, Verbose: true, Markdown: false, System: "Sys"}

    mergeFileConfig(cfg, fc)

    assert.Equal(t, "openai", cfg.Provider)
    assert.Equal(t, "gpt-4", cfg.Model)
    assert.Equal(t, 0.8, cfg.Temperature)
    assert.Equal(t, true, cfg.Verbose)      // only set when true
    assert.Equal(t, false, cfg.Markdown)    // only set when false
    assert.Equal(t, "Sys", cfg.System)

    // When file has zero-values that shouldn't override defaults
    cfg2 := &Config{Provider: "copilot", Model: "m", Temperature: 0.6, Verbose: true, Markdown: false, System: "S"}
    fc2 := &FileConfig{} // all zero
    mergeFileConfig(cfg2, fc2)
    assert.Equal(t, true, cfg2.Verbose)
    assert.Equal(t, false, cfg2.Markdown)
}

func TestSetProviderConfig_FileAndEnvPrecedence(t *testing.T) {
    tmp := t.TempDir()
    os.Setenv("XDG_CONFIG_HOME", tmp)
    defer os.Unsetenv("XDG_CONFIG_HOME")

    // Ensure a clean env so defaults/fixtures are deterministic.
    os.Unsetenv("OPENAI_API_KEY")
    os.Unsetenv("OPENAI_API_BASE")
    os.Unsetenv("COPILOT_API_KEY")
    os.Unsetenv("COPILOT_API_BASE")
    os.Unsetenv("GEMINI_API_KEY")
    os.Unsetenv("GEMINI_API_BASE")

    // Create YAML with openai provider credentials
    dir := filepath.Join(tmp, "gpt-cli")
    _ = os.MkdirAll(dir, 0o755)
    yaml := "providers:\n  openai:\n    api_key: file-key\n    base_url: https://file-base\n"
    _ = os.WriteFile(filepath.Join(dir, "config.yml"), []byte(yaml), 0o644)

    // From file
    cfg := &Config{Provider: "openai"}
    setProviderConfig(cfg)
    assert.Equal(t, "file-key", cfg.APIKey)
    assert.Equal(t, "https://file-base", cfg.BaseURL)

    // Env overrides file
    os.Setenv("OPENAI_API_KEY", "env-key")
    os.Setenv("OPENAI_API_BASE", "https://env-base")
    defer os.Unsetenv("OPENAI_API_KEY")
    defer os.Unsetenv("OPENAI_API_BASE")
    cfg2 := &Config{Provider: "openai"}
    setProviderConfig(cfg2)
    assert.Equal(t, "env-key", cfg2.APIKey)
    assert.Equal(t, "https://env-base", cfg2.BaseURL)

    // Default base when none provided for openai
    os.Unsetenv("OPENAI_API_BASE")
    _ = os.WriteFile(filepath.Join(dir, "config.yml"), []byte("providers:\n  openai:\n    api_key: file-key\n"), 0o644)
    cfg3 := &Config{Provider: "openai"}
    setProviderConfig(cfg3)
    assert.Equal(t, "https://api.openai.com", cfg3.BaseURL)

    // Gemini default base when none provided
    os.Unsetenv("GEMINI_API_BASE")
    _ = os.WriteFile(filepath.Join(dir, "config.yml"), []byte("providers:\n  gemini:\n    api_key: g\n"), 0o644)
    cfg4 := &Config{Provider: "gemini"}
    setProviderConfig(cfg4)
    assert.Equal(t, "https://generativelanguage.googleapis.com/v1beta/openai", cfg4.BaseURL)

    // Copilot has no default base
    _ = os.WriteFile(filepath.Join(dir, "config.yml"), []byte("providers:\n  copilot:\n    api_key: c\n"), 0o644)
    cfg5 := &Config{Provider: "copilot"}
    setProviderConfig(cfg5)
    assert.Equal(t, "", cfg5.BaseURL)
}
