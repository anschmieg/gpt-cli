//go:build integration
// +build integration

package modes

import (
    "bytes"
    "os"
    "testing"
    "github.com/anschmieg/gpt-cli/internal/config"
    "github.com/anschmieg/gpt-cli/internal/ui"
    "github.com/stretchr/testify/assert"
)

type integProv struct{ resp string }
func (p integProv) GetName() string { return "mock" }
func (p integProv) CallProvider(prompt string) (string, error) { return p.resp, nil }
func (p integProv) StreamProvider(prompt string) (<-chan string, <-chan error) { c:=make(chan string); e:=make(chan error); close(c); close(e); return c,e }

func captureStdout(f func()) string {
    old := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w
    f()
    _ = w.Close()
    var buf bytes.Buffer
    _, _ = buf.ReadFrom(r)
    os.Stdout = old
    return buf.String()
}

func TestShellMode_Interactive_Integration(t *testing.T) {
    cfg := &config.Config{Provider: "mock", Model: "m"}
    u := ui.New()
    p := integProv{resp: `{"command":"echo hi","safety_level":"safe","explanation":"ok"}`}
    m := NewShellMode(cfg, p, u)

    // simulate abort input to avoid blocking
    oldIn := os.Stdin
    rin, win, _ := os.Pipe()
    os.Stdin = rin
    defer func(){ os.Stdin = oldIn }()
    go func(){ _, _ = win.Write([]byte("a\n")); _ = win.Close() }()

    out := captureStdout(func(){ _ = m.InteractiveMode("say hi") })
    assert.Contains(t, out, "Command Suggestion")
    assert.Contains(t, out, "echo hi")
}

func TestChatMode_InitWithInitialPrompt_Integration(t *testing.T) {
    cfg := &config.Config{Provider: "mock", Model: "m"}
    u := ui.New()
    p := integProv{resp: "Hello from chat"}
    mode := NewChatMode(cfg, p, u)
    m := NewChatModel(mode, "hi")
    cmd := m.Init()
    if cmd == nil { t.Fatalf("expected cmd from Init") }
    // Resolve request
    msg := cmd()
    _, _ = m.Update(msg)
    // After response, conversation should include assistant message
    if len(m.chatMode.conversation.Messages) == 0 {
        t.Fatalf("expected messages")
    }
}
