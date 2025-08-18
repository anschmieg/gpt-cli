//go:build integration
// +build integration

package app

import (
    "testing"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/stretchr/testify/assert"
)

type mockProv struct{ out string; err error }
func (m mockProv) GetName() string { return "mock" }
func (m mockProv) CallProvider(prompt string) (string, error) { return m.out, m.err }
func (m mockProv) StreamProvider(prompt string) (<-chan string, <-chan error) {
    c := make(chan string); e := make(chan error); close(c); close(e); return c, e
}

// Drives the model through input -> loading -> response using a fake provider.
func TestModel_EndToEnd_NoNetwork(t *testing.T) {
    m := NewModel()
    // inject mock provider
    m.provider = mockProv{out: "Hello"}

    // Type "hi" and press enter
    _, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
    _, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
    _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
    assert.NotNil(t, cmd)

    // Resolve command
    msg := cmd()
    _, _ = m.Update(msg)
    assert.Equal(t, StateResponse, m.State())
}

