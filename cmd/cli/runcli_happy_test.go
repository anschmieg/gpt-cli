package main

import (
    "bytes"
    "fmt"
    "os"
    "testing"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/stretchr/testify/assert"
)

func captureStdoutLocal(f func()) string {
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

func TestRunCLI_NoArgs_ShowsHelp_AndNoStreamOverrides(t *testing.T) {
    // Reset globals
    shellMode, chatMode = false, false
    stream, noStream = true, false

    // Use only --no-stream flag; no args -> help prints
    rootCmd.SetArgs([]string{"--no-stream"})
    out := captureStdoutLocal(func() {
        _ = rootCmd.Execute()
    })

    // Verify stream override applied and help text printed
    assert.False(t, stream)
    assert.Contains(t, out, "gpt-cli is a command-line interface")
}

func TestRunCLI_ChatMode_ErrorExitCoversBranch(t *testing.T) {
    // Reset globals
    shellMode, chatMode = false, false
    stream, noStream = true, false

    // Override program runner to simulate BubbleTea run error quickly
    prevProg := newProgram
    newProgram = func(m tea.Model) programRunner { return fakeErrRunner{} }
    defer func() { newProgram = prevProg }()

    var gotExit int
    prevExit := exitFunc
    exitFunc = func(code int) { gotExit = code }
    defer func() { exitFunc = prevExit }()

    // --chat with no args triggers runChatMode which will fail in sandbox and exit
    rootCmd.SetArgs([]string{"--chat"})
    errOut := captureStderr(func() {
        _ = rootCmd.Execute()
    })

    assert.Equal(t, 1, gotExit)
    assert.Contains(t, errOut, "Error running chat mode:")
}

// fakeErrRunner implements programRunner for tests
type fakeErrRunner struct{}
func (fakeErrRunner) Run() (tea.Model, error) { return nil, fmt.Errorf("unit test error") }

func TestRunCLI_ShellMode_ProviderErrorExit(t *testing.T) {
    // Ensure environment doesn't provide COPILOT_API_BASE
    os.Unsetenv("COPILOT_API_BASE")

    // Reset globals
    shellMode, chatMode = false, false
    stream, noStream = true, false

    // Defensive: if provider unexpectedly succeeds in this environment,
    // stub stdin to immediately abort interactive mode to avoid hanging.
    oldIn := os.Stdin
    inR, inW, _ := os.Pipe()
    os.Stdin = inR
    defer func() { os.Stdin = oldIn }()
    go func() {
        _, _ = inW.Write([]byte("a\n"))
        _ = inW.Close()
    }()

    var gotExit int
    prevExit := exitFunc
    exitFunc = func(code int) { gotExit = code }
    defer func() { exitFunc = prevExit }()

    // Shell mode with provider copilot but without base URL triggers provider error
    rootCmd.SetArgs([]string{"--shell", "--provider", "copilot", "list files"})
    errOut := captureStderr(func() {
        _ = rootCmd.Execute()
    })

    assert.Equal(t, 1, gotExit)
    assert.Contains(t, errOut, "Error in shell mode:")
}

func TestRunCLI_NonInteractive_StreamingErrorExit(t *testing.T) {
    // Ensure env doesn't provide base
    os.Unsetenv("COPILOT_API_BASE")
    os.Unsetenv("COPILOT_API_KEY")

    shellMode, chatMode = false, false
    stream, noStream = true, false

    exitCode := 0
    prevExit := exitFunc
    exitFunc = func(code int) { exitCode = code }
    defer func() { exitFunc = prevExit }()

    rootCmd.SetArgs([]string{"--provider", "copilot", "hello"})
    errOut := captureStderr(func() { _ = rootCmd.Execute() })

    assert.Equal(t, 1, exitCode)
    assert.Contains(t, errOut, "Error:")
}

func TestRunCLI_NonInteractive_NoStreamErrorExit(t *testing.T) {
    // Ensure env doesn't provide base
    os.Unsetenv("COPILOT_API_BASE")
    os.Unsetenv("COPILOT_API_KEY")

    shellMode, chatMode = false, false
    stream, noStream = true, false

    exitCode := 0
    prevExit := exitFunc
    exitFunc = func(code int) { exitCode = code }
    defer func() { exitFunc = prevExit }()

    rootCmd.SetArgs([]string{"--no-stream", "--provider", "copilot", "hello"})
    errOut := captureStderr(func() { _ = rootCmd.Execute() })

    assert.Equal(t, 1, exitCode)
    assert.Contains(t, errOut, "Error:")
}

func TestRunCLI_NoArgs_EnvOverridesPath(t *testing.T) {
    // Set env vars that runCLI reads when flags not explicitly changed
    os.Setenv("GPT_CLI_PROVIDER", "openai")
    os.Setenv("GPT_CLI_MODEL", "gpt-4")
    os.Setenv("GPT_CLI_TEMPERATURE", "0.7")
    defer os.Unsetenv("GPT_CLI_PROVIDER")
    defer os.Unsetenv("GPT_CLI_MODEL")
    defer os.Unsetenv("GPT_CLI_TEMPERATURE")

    shellMode, chatMode = false, false
    stream, noStream = true, false

    rootCmd.SetArgs([]string{})
    out := captureStdoutLocal(func() { _ = rootCmd.Execute() })
    assert.Contains(t, out, "Usage:")
}

func TestRunCLI_FlagsChangedPath(t *testing.T) {
    shellMode, chatMode = false, false
    stream, noStream = true, false

    rootCmd.SetArgs([]string{"--provider", "openai", "--model", "gpt-4o-mini", "--temperature", "0.5"})
    out := captureStdoutLocal(func() { _ = rootCmd.Execute() })
    assert.Contains(t, out, "flags") // help includes flags listing
}

func TestRunCLI_NoArgs_GeminiEnvProviderPath(t *testing.T) {
    os.Setenv("GPT_CLI_PROVIDER", "gemini")
    defer os.Unsetenv("GPT_CLI_PROVIDER")

    shellMode, chatMode = false, false
    stream, noStream = true, false

    rootCmd.SetArgs([]string{})
    out := captureStdoutLocal(func() { _ = rootCmd.Execute() })
    assert.Contains(t, out, "Usage:")
}
