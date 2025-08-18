package main

import (
    "bytes"
    "errors"
    "os"
    "testing"
    "time"

    "github.com/anschmieg/gpt-cli/internal/config"
    "github.com/anschmieg/gpt-cli/internal/utils"
    "github.com/stretchr/testify/assert"
)

type errStreamProvider struct{}

func (errStreamProvider) CallProvider(prompt string) (string, error) { return "", errors.New("call failed") }
func (errStreamProvider) StreamProvider(prompt string) (<-chan string, <-chan error) {
    c := make(chan string)      // left open so select can receive from error
    e := make(chan error, 1)
    e <- errors.New("stream failed")
    close(e)
    return c, e
}
func (errStreamProvider) GetName() string { return "mock" }

// captureStderr is also defined in runcli_test.go; redefine here to keep tests independent.
func captureStderrLocal(f func()) string {
    old := os.Stderr
    r, w, _ := os.Pipe()
    os.Stderr = w
    f()
    w.Close()
    var buf bytes.Buffer
    _, _ = buf.ReadFrom(r)
    os.Stderr = old
    return buf.String()
}

func TestRunNonInteractiveWithProvider_StreamingErrorExits(t *testing.T) {
    cfg := &config.Config{Markdown: true}
    logger := utils.NewLogger(false)
    p := errStreamProvider{}

    exitCh := make(chan int, 1)
    prevExit := exitFunc
    exitFunc = func(code int) { exitCh <- code }
    defer func() { exitFunc = prevExit }()

    done := make(chan string, 1)
    go func() {
        stderr := captureStderrLocal(func() {
            runNonInteractiveWithProvider(cfg, "prompt", logger, p, true)
        })
        done <- stderr
    }()

    select {
    case code := <-exitCh:
        assert.Equal(t, 1, code)
    case <-time.After(200 * time.Millisecond):
        t.Fatal("exit not called")
    }

    // We can't guarantee goroutine returns (loop blocks); but if it does, verify stderr
    select {
    case stderr := <-done:
        assert.Contains(t, stderr, "stream failed")
    case <-time.After(50 * time.Millisecond):
        // acceptable
    }
}

func TestRunNonInteractiveWithProvider_NonStreamingErrorExits(t *testing.T) {
    cfg := &config.Config{Markdown: true}
    logger := utils.NewLogger(false)
    p := errStreamProvider{}

    var gotExit int
    prevExit := exitFunc
    exitFunc = func(code int) { gotExit = code }
    defer func() { exitFunc = prevExit }()

    stderr := captureStderrLocal(func() {
        runNonInteractiveWithProvider(cfg, "prompt", logger, p, false)
    })

    assert.Equal(t, 1, gotExit)
    assert.Contains(t, stderr, "call failed")
}

func TestStreamingMarkdownBranch(t *testing.T) {
    cfg := &config.Config{Markdown: true}
    mock := &MockStreamProvider{chunks: []string{"# Head\n", "\n", "Body"}}
    logger := utils.NewLogger(false)

    out := captureStdout(func() { runNonInteractiveWithProvider(cfg, "prompt", logger, mock, true) })
    // Streaming + markdown branch should include header text
    assert.Contains(t, out, "Head")
    assert.NotContains(t, out, "\n\n\n")
}
