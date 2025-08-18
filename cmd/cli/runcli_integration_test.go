//go:build integration
// +build integration

package main

import (
    "bytes"
    "os"
    "testing"

    "github.com/anschmieg/gpt-cli/internal/config"
    "github.com/anschmieg/gpt-cli/internal/utils"
    "github.com/stretchr/testify/assert"
)

// minimal mock provider for integration tests
type integProvider struct{ chunks []string; err error }
func (p integProvider) GetName() string { return "mock" }
func (p integProvider) CallProvider(prompt string) (string, error) {
    if p.err != nil { return "", p.err }
    var s string
    for _, c := range p.chunks { s += c }
    return s, nil
}
func (p integProvider) StreamProvider(prompt string) (<-chan string, <-chan error) {
    c := make(chan string)
    e := make(chan error)
    go func(){
        defer close(c); defer close(e)
        for _, s := range p.chunks { c <- s }
    }()
    return c, e
}

func captureStdoutInteg(f func()) string {
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

func TestCLI_RunNonInteractive_WithMockProvider_StreamingAndNonStreaming(t *testing.T) {
    cfg := &config.Config{Markdown: true, Provider: "mock", Model: "m"}
    mock := integProvider{chunks: []string{"# H", "ello"}}
    logger := utils.NewLogger(false)

    // Non-streaming path renders markdown
    out := captureStdoutInteg(func(){ runNonInteractiveWithProvider(cfg, "p", logger, mock, false) })
    assert.Contains(t, out, "H")

    // Streaming path prints incrementally
    out2 := captureStdoutInteg(func(){ runNonInteractiveWithProvider(cfg, "p", logger, mock, true) })
    assert.NotEmpty(t, out2)
}

