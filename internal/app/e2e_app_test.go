//go:build e2e
// +build e2e

package app

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/anschmieg/gpt-cli/internal/providers"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestE2E_AppModel_RequestFlow(t *testing.T) {
    providers.E2ECallResponder = func(name, prompt string) (string, error) {
        return "resp:" + prompt, nil
    }
    m := NewModel()
    // type 'x' and press enter
    _, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
    _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
    msg := cmd()
    _, _ = m.Update(msg)
    assert.Equal(t, StateResponse, m.State())
}

