package main

import (
    "bytes"
    "os"
    "testing"

    "github.com/stretchr/testify/assert"
)

// captureStderr runs f while redirecting stderr, returning captured output.
func captureStderr(f func()) string {
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

func TestRunCLI_ConflictingModes(t *testing.T) {
    // Intercept exit
    var gotExit int
    prevExit := exitFunc
    exitFunc = func(code int) { gotExit = code }
    defer func() { exitFunc = prevExit }()

    // Ensure clean flag state
    shellMode, chatMode = false, false
    stream, noStream = true, false
    // Prepare args: both --shell and --chat with a prompt
    rootCmd.SetArgs([]string{"--shell", "--chat", "hello"})
    stderr := captureStderr(func() {
        _ = rootCmd.Execute()
    })

    assert.Equal(t, 1, gotExit)
    assert.Contains(t, stderr, "Cannot use both --shell and --chat")
}

func TestRunCLI_ShellModeMissingPrompt(t *testing.T) {
    var gotExit int
    prevExit := exitFunc
    exitFunc = func(code int) { gotExit = code }
    defer func() { exitFunc = prevExit }()

    // Ensure clean flag state
    shellMode, chatMode = false, false
    stream, noStream = true, false
    // Shell mode without a prompt should print error + usage and exit(1)
    rootCmd.SetArgs([]string{"--shell"})
    stderr := captureStderr(func() {
        _ = rootCmd.Execute()
    })

    assert.Equal(t, 1, gotExit)
    assert.Contains(t, stderr, "Shell mode requires a prompt")
}
