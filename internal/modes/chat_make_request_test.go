package modes

import (
    "errors"
    "testing"

    "github.com/anschmieg/gpt-cli/internal/config"
    "github.com/anschmieg/gpt-cli/internal/ui"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/stretchr/testify/assert"
)

type cmdMockProvider struct{ resp string; err error }

func (m cmdMockProvider) CallProvider(prompt string) (string, error) { if m.err != nil { return "", m.err }; return m.resp, nil }
func (m cmdMockProvider) StreamProvider(prompt string) (<-chan string, <-chan error) { c := make(chan string); e := make(chan error); close(c); close(e); return c, e }
func (m cmdMockProvider) GetName() string { return "mock" }

func runTeaCmd(cmd tea.Cmd) tea.Msg { return cmd() }

func TestMakeRequest_SuccessAndError(t *testing.T) {
    cfg := &config.Config{Provider: "mock", Model: "m"}
    u := ui.New()

    // Success
    cm := NewChatMode(cfg, cmdMockProvider{resp: "ok"}, u)
    model := NewChatModel(cm, "")
    msg := runTeaCmd(model.makeRequest())
    rm, ok := msg.(ResponseMsg)
    assert.True(t, ok)
    assert.Equal(t, "ok", rm.Content)

    // Error
    cm2 := NewChatMode(cfg, cmdMockProvider{err: errors.New("boom")}, u)
    model2 := NewChatModel(cm2, "")
    msg2 := runTeaCmd(model2.makeRequest())
    em, ok := msg2.(ErrorMsg)
    assert.True(t, ok)
    assert.Contains(t, em.Error, "boom")
}

