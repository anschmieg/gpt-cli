//go:build e2e
// +build e2e

package modes

import (
    "testing"
    "github.com/anschmieg/gpt-cli/internal/config"
    "github.com/anschmieg/gpt-cli/internal/providers"
    "github.com/anschmieg/gpt-cli/internal/ui"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/stretchr/testify/assert"
)

func TestE2E_ChatMode_SendMessage(t *testing.T) {
    providers.E2ECallResponder = func(name, prompt string) (string, error) {
        return "ok", nil
    }
    cfg := &config.Config{Provider: "gemini", Model: "m"}
    // obtain provider through API-managing module
    p := providers.NewProvider(cfg.Provider, cfg)
    u := ui.New()
    mode := NewChatMode(cfg, p, u)
    model := NewChatModel(mode, "")
    // type 'hi' and enter
    _, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
    _, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
    _, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
    // resolve request
    msg := cmd()
    _, _ = model.Update(msg)
    assert.Equal(t, ChatStateResponse, model.state)
}

