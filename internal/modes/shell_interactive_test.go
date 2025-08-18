package modes

import (
    "bytes"
    "os"
    "testing"

    "github.com/anschmieg/gpt-cli/internal/config"
    "github.com/anschmieg/gpt-cli/internal/ui"
    "github.com/stretchr/testify/assert"
)

// Provides a minimal valid JSON suggestion for InteractiveMode to render.
type interactiveMockProvider struct{}

func (interactiveMockProvider) CallProvider(prompt string) (string, error) {
    return `{"command":"echo hi","safety_level":"safe","explanation":"prints","reasoning":"readonly"}`, nil
}
func (interactiveMockProvider) StreamProvider(prompt string) (<-chan string, <-chan error) {
    ch := make(chan string, 1)
    errc := make(chan error, 1)
    close(ch)
    close(errc)
    return ch, errc
}
func (interactiveMockProvider) GetName() string { return "mock" }

func TestInteractiveMode_AbortPath(t *testing.T) {
    cfg := &config.Config{Provider: "mock", Model: "m"}
    p := interactiveMockProvider{}
    u := ui.New()
    mode := NewShellMode(cfg, p, u)

    // Simulate typing 'a' (abort) then newline at the prompt using a pipe as stdin.
    oldStdin := os.Stdin
    rIn, wIn, _ := os.Pipe()
    os.Stdin = rIn
    defer func() { os.Stdin = oldStdin }()

    // Capture output to ensure it prints suggestion content
    var out bytes.Buffer
    oldStdout := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w

    // Write the input concurrently then close write end
    go func() {
        _, _ = wIn.Write([]byte("a\n"))
        _ = wIn.Close()
    }()

    err := mode.InteractiveMode("list files")

    w.Close()
    _, _ = out.ReadFrom(r)
    os.Stdout = oldStdout

    assert.NoError(t, err)
    // Should include section header and fields from suggestion
    assert.Contains(t, out.String(), "Command Suggestion")
    assert.Contains(t, out.String(), "echo hi")
    assert.Contains(t, out.String(), "SAFE") // safety level uppercased by renderer
}

// no-op
