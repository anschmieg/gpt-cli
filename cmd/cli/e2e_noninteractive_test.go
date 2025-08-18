//go:build e2e
// +build e2e

package main

import (
    "bytes"
    "os"
    "testing"

    "github.com/anschmieg/gpt-cli/internal/config"
    "github.com/anschmieg/gpt-cli/internal/providers"
    "github.com/anschmieg/gpt-cli/internal/utils"
    "github.com/stretchr/testify/assert"
)

func capStdout(f func()) string {
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

func TestE2E_NonInteractive_NonStreaming_PlainAndMarkdown(t *testing.T) {
    // Mock provider behavior: echo back the prompt with marker
    providers.E2ECallResponder = func(name, prompt string) (string, error) {
        if prompt == "md" { return "# Header\n\nBody", nil }
        return "plain text", nil
    }

    cfg := &config.Config{Provider: "openai", Model: "m", Markdown: false}
    logger := utils.NewLogger(false)

    // Not TTY should print plain text even if markdown true is ignored here
    prevTTY := isTerminalFunc
    isTerminalFunc = func(fd uintptr) bool { return false }
    out := capStdout(func(){ runNonInteractive(cfg, "plain", logger) })
    isTerminalFunc = prevTTY
    assert.Contains(t, out, "plain text")

    // When TTY and markdown true, renderer should be used
    cfg.Markdown = true
    isTerminalFunc = func(fd uintptr) bool { return true }
    out2 := capStdout(func(){ runNonInteractive(cfg, "md", logger) })
    isTerminalFunc = prevTTY
    assert.Contains(t, out2, "Header")
}

func TestE2E_NonInteractive_Streaming_WithChunks(t *testing.T) {
    providers.E2EStreamResponder = func(name, prompt string) (<-chan string, <-chan error) {
        c := make(chan string, 3)
        e := make(chan error, 1)
        c <- "Hello"
        c <- " "
        c <- "World"
        close(c); close(e)
        return c, e
    }
    cfg := &config.Config{Provider: "openai", Model: "m", Markdown: false}
    logger := utils.NewLogger(false)

    // Enable streaming
    stream, noStream = true, false
    out := capStdout(func(){ runNonInteractive(cfg, "ignored", logger) })
    assert.Contains(t, out, "Hello World")
}

