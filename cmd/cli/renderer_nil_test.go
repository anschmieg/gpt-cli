package main

import (
    "testing"

    "github.com/anschmieg/gpt-cli/internal/config"
    "github.com/anschmieg/gpt-cli/internal/ui"
    "github.com/anschmieg/gpt-cli/internal/utils"
    "github.com/stretchr/testify/assert"
)

// nilRenderer returns a UI with a nil Renderer to exercise the markdown
// fallback branch in runNonInteractiveWithProvider.
func nilRenderer() *ui.UI {
    u := ui.New()
    u.Renderer = nil
    return u
}

func TestNonStreamingMarkdownFallbackWhenRendererNil(t *testing.T) {
    // Override UI factory to simulate nil renderer
    prev := newUI
    newUI = nilRenderer
    defer func() { newUI = prev }()

    // Force TTY path
    prevTTY := isTerminalFunc
    isTerminalFunc = func(fd uintptr) bool { return true }
    defer func() { isTerminalFunc = prevTTY }()

    cfg := &config.Config{Markdown: true}
    mock := &MockStreamProvider{chunks: []string{"Hello World"}}
    logger := utils.NewLogger(false)

    out := captureStdout(func() { runNonInteractiveWithProvider(cfg, "prompt", logger, mock, false) })
    // The fallback path for non-streaming prints plain response
    assert.Contains(t, out, "Hello World")
}
