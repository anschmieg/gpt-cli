package main

import (
    "strings"
    "testing"

    "github.com/anschmieg/gpt-cli/internal/config"
    "github.com/anschmieg/gpt-cli/internal/utils"
    "github.com/stretchr/testify/assert"
)

func TestStreamingMarkdownTrickyFragments(t *testing.T) {
    cfg := &config.Config{Markdown: true}
    // Split a code fence and include blank lines/lists/quotes across chunks
    chunks := []string{
        "# Title\n\nIntro\n```\ncode line 1\n",
        "code line 2\n```\n\n- item one\n",
        "\n> quoted\n",
    }

    mock := &MockStreamProvider{chunks: chunks}
    logger := utils.NewLogger(false)

    // Force tty
    prevIsTTY := isTerminalFunc
    isTerminalFunc = func(fd uintptr) bool { return true }
    defer func() { isTerminalFunc = prevIsTTY }()

    out := captureStdout(func() { runNonInteractiveWithProvider(cfg, "prompt", logger, mock, true) })

    // Basic smoke assertions: important tokens present and no excessive blank runs
    assert.Contains(t, out, "Title")
    assert.Contains(t, out, "Intro")
    assert.Contains(t, out, "code line 1")
    assert.Contains(t, out, "code line 2")
    assert.Contains(t, out, "item one")
    assert.Contains(t, out, "quoted")
    // no 3+ consecutive newlines
    assert.NotContains(t, out, "\n\n\n")
}

func TestNonStreamingMarkdownTricky(t *testing.T) {
    cfg := &config.Config{Markdown: true}
    body := strings.Join([]string{
        "# T\n",
        "\n",
        "```\n",
        "a\n",
        "b\n",
        "```\n",
        "\n",
        "- x\n",
        "> y\n",
        "",
    }, "")

    mock := &MockStreamProvider{chunks: []string{body}}
    logger := utils.NewLogger(false)

    prevIsTTY := isTerminalFunc
    isTerminalFunc = func(fd uintptr) bool { return true }
    defer func() { isTerminalFunc = prevIsTTY }()

    out := captureStdout(func() { runNonInteractiveWithProvider(cfg, "prompt", logger, mock, false) })
    assert.Contains(t, out, "T")
    assert.Contains(t, out, "a")
    assert.Contains(t, out, "b")
    assert.Contains(t, out, "x")
    assert.Contains(t, out, "y")
}

