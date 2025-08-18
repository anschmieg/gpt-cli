package modes

import (
    "bytes"
    "os"
    "testing"

    "github.com/anschmieg/gpt-cli/internal/config"
    "github.com/anschmieg/gpt-cli/internal/ui"
    "github.com/stretchr/testify/assert"
)

func TestExecuteCommandEcho(t *testing.T) {
    s := &ShellMode{config: &config.Config{}, ui: ui.New()}

    // Capture stdout
    old := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w

    err := s.executeCommand("echo hello")

    w.Close()
    var buf bytes.Buffer
    _, _ = buf.ReadFrom(r)
    os.Stdout = old

    assert.NoError(t, err)
    assert.Contains(t, buf.String(), "hello")
}

func TestEditAndExecute_NoChangeUsesOriginal(t *testing.T) {
    s := &ShellMode{config: &config.Config{}, ui: ui.New()}

    // Provide empty edited input so it falls back to original
    inR, inW, _ := os.Pipe()
    oldIn := os.Stdin
    os.Stdin = inR
    defer func() { os.Stdin = oldIn }()

    // Capture stdout
    old := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w

    go func() {
        _, _ = inW.Write([]byte("\n"))
        _ = inW.Close()
    }()

    err := s.editAndExecute("echo edited")

    w.Close()
    var buf bytes.Buffer
    _, _ = buf.ReadFrom(r)
    os.Stdout = old

    assert.NoError(t, err)
    assert.Contains(t, buf.String(), "edited")
}

func TestPromptUserAction_ExecutePath(t *testing.T) {
    s := &ShellMode{config: &config.Config{}, ui: ui.New()}
    sugg := &ShellSuggestion{Command: "echo via-prompt", SafetyLevel: "safe", Explanation: "e"}

    // Feed 'e\n' to choose execute path
    inR, inW, _ := os.Pipe()
    oldIn := os.Stdin
    os.Stdin = inR
    defer func() { os.Stdin = oldIn }()

    // Capture stdout
    old := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w

    go func() {
        _, _ = inW.Write([]byte("e\n"))
        _ = inW.Close()
    }()

    err := s.promptUserAction(sugg)

    w.Close()
    var buf bytes.Buffer
    _, _ = buf.ReadFrom(r)
    os.Stdout = old

    assert.NoError(t, err)
    assert.Contains(t, buf.String(), "Executing:")
    assert.Contains(t, buf.String(), "via-prompt")
}

type refineProvider struct{ refined string }
func (refineProvider) GetName() string { return "mock" }
func (r refineProvider) CallProvider(prompt string) (string, error) { return r.refined, nil }
func (refineProvider) StreamProvider(prompt string) (<-chan string, <-chan error) { c := make(chan string); e := make(chan error); close(c); close(e); return c, e }

func TestPromptUserAction_RefinePathThenAbort(t *testing.T) {
    s := &ShellMode{config: &config.Config{}, ui: ui.New(), provider: refineProvider{refined: `{"command":"echo refined","safety_level":"safe","explanation":"ok"}`}}
    sugg := &ShellSuggestion{Command: "echo base", SafetyLevel: "safe", Explanation: "e"}

    // Feed: r (refine) -> "better" -> a (abort)
    inR, inW, _ := os.Pipe()
    oldIn := os.Stdin
    os.Stdin = inR
    defer func() { os.Stdin = oldIn }()

    // Capture stdout
    old := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w

    go func() {
        _, _ = inW.Write([]byte("r\n"))
        _, _ = inW.Write([]byte("better\n"))
        _, _ = inW.Write([]byte("a\n"))
        _ = inW.Close()
    }()

    err := s.promptUserAction(sugg)

    w.Close()
    var buf bytes.Buffer
    _, _ = buf.ReadFrom(r)
    os.Stdout = old

    assert.NoError(t, err)
    // Should include refined command once refinement is applied
    assert.Contains(t, buf.String(), "echo refined")
}

