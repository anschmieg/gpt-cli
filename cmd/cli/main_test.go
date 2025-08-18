package main

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/anschmieg/gpt-cli/internal/config"
	"github.com/anschmieg/gpt-cli/internal/utils"
	"github.com/stretchr/testify/assert"
)

// MockProvider for tests
type MockStreamProvider struct {
	chunks []string
	err    error
}

func (m *MockStreamProvider) CallProvider(prompt string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	// join chunks for non-streaming call
	result := ""
	for _, c := range m.chunks {
		result += c
	}
	return result, nil
}

func (m *MockStreamProvider) StreamProvider(prompt string) (<-chan string, <-chan error) {
	c := make(chan string)
	e := make(chan error)

	go func() {
		defer close(c)
		defer close(e)
		for _, chunk := range m.chunks {
			c <- chunk
			time.Sleep(10 * time.Millisecond)
		}
	}()

	return c, e
}

func (m *MockStreamProvider) GetName() string { return "mock" }

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = old
	return buf.String()
}

func TestNonStreamingRendersPlainWhenNotTTY(t *testing.T) {
	cfg := &config.Config{Markdown: true}
	mock := &MockStreamProvider{chunks: []string{"# Header\n\nHello"}}
	logger := utils.NewLogger(false)

	// Force non-tty
	prevIsTTY := isTerminalFunc
	isTerminalFunc = func(fd uintptr) bool { return false }
	defer func() { isTerminalFunc = prevIsTTY }()

	out := captureStdout(func() { runNonInteractiveWithProvider(cfg, "prompt", logger, mock, false) })
	// Should output raw markdown text (not rendered) because not a TTY
	assert.Contains(t, out, "# Header")
}

func TestNonStreamingRendersMarkdownWhenTTY(t *testing.T) {
	cfg := &config.Config{Markdown: true}
	mock := &MockStreamProvider{chunks: []string{"# Header\n\nHello"}}
	logger := utils.NewLogger(false)

	// Force tty
	prevIsTTY := isTerminalFunc
	isTerminalFunc = func(fd uintptr) bool { return true }
	defer func() { isTerminalFunc = prevIsTTY }()

	out := captureStdout(func() { runNonInteractiveWithProvider(cfg, "prompt", logger, mock, false) })
	// Rendered markdown should contain the header text
	assert.Contains(t, out, "Header")
}

func TestStreamingPrintsChunks(t *testing.T) {
	cfg := &config.Config{Markdown: false}
	mock := &MockStreamProvider{chunks: []string{"Hello", " ", "World"}}
	logger := utils.NewLogger(false)

	out := captureStdout(func() { runNonInteractiveWithProvider(cfg, "prompt", logger, mock, true) })
	assert.Contains(t, out, "Hello World")
}

func TestNonStreamingPlainWhenMarkdownDisabled(t *testing.T) {
    cfg := &config.Config{Markdown: false}
    mock := &MockStreamProvider{chunks: []string{"# Header\n\nHello"}}
    logger := utils.NewLogger(false)

    out := captureStdout(func() { runNonInteractiveWithProvider(cfg, "prompt", logger, mock, false) })
    // Plain output should preserve markdown syntax
    assert.Contains(t, out, "# Header")
}

func TestJoinArgs(t *testing.T) {
    got := joinArgs([]string{"one", "two", "three"})
    assert.Equal(t, "one two three", got)
    assert.Equal(t, "", joinArgs(nil))
}

func TestGetEnvOrDefault_CLI(t *testing.T) {
    const k = "TEST_CLI_ENV_FN"
    // Not set -> default
    os.Unsetenv(k)
    assert.Equal(t, "d", getEnvOrDefault(k, "d"))
    // Set -> value
    os.Setenv(k, "v")
    defer os.Unsetenv(k)
    assert.Equal(t, "v", getEnvOrDefault(k, "d"))
}
