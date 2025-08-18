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

func capStdoutE2E(f func()) string {
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

func TestE2E_ShellMode_SuggestAndAbort(t *testing.T) {
    // Provider returns a valid suggestion JSON
    providers.E2ECallResponder = func(name, prompt string) (string, error) {
        return `{"command":"echo hi","safety_level":"safe","explanation":"prints"}`, nil
    }

    cfg := &config.Config{Provider: "copilot", Model: "m"}
    logger := utils.NewLogger(false)

    // Simulate user input 'a' to abort after printing suggestion
    oldIn := os.Stdin
    rin, win, _ := os.Pipe()
    os.Stdin = rin
    defer func(){ os.Stdin = oldIn }()
    go func(){ _, _ = win.Write([]byte("a\n")); _ = win.Close() }()

    out := capStdoutE2E(func(){ runShellMode(cfg, "list", logger) })
    assert.Contains(t, out, "Command Suggestion")
    assert.Contains(t, out, "echo hi")
}

