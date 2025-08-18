package main

import (
    "bytes"
    "encoding/json"
    "io"
    "net/http"
    "net/http/httptest"
    "os"
    "os/exec"
    "testing"
    "strings"
)

// mockOpenAI creates a test server that mimics minimal OpenAI responses.
func mockOpenAI() *httptest.Server {
    var lastPrompt string
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var req struct {
            Messages []struct {
                Role    string `json:"role"`
                Content string `json:"content"`
            } `json:"messages"`
        }
        _ = json.NewDecoder(r.Body).Decode(&req)
        user := ""
        for i := len(req.Messages)-1; i >=0; i-- {
            if req.Messages[i].Role == "user" {
                user = strings.ToLower(req.Messages[i].Content)
                break
            }
        }
        var content string
        switch {
        case strings.Contains(user, "list files"):
            content = `{"command":"ls -la","safety_level":"safe","explanation":"Lists files","reasoning":"Read-only"}`
        case strings.Contains(user, "who am i"):
            content = "You previously said: " + lastPrompt
        case strings.Contains(user, "hello"):
            content = "Hi there!"
        case strings.Contains(user, "how are you"):
            content = "I'm good."
        case strings.Contains(user, "2+2"):
            content = "4"
        default:
            content = "default"
        }
        lastPrompt = user
        resp := map[string]any{
            "choices": []map[string]map[string]string{
                {"message": {"content": content}},
            },
        }
        _ = json.NewEncoder(w).Encode(resp)
    })
    return httptest.NewServer(handler)
}

func runCLI(t *testing.T, serverURL string, args ...string) (string, error) {
    cmd := exec.Command("go", append([]string{"run", "./cmd/cli", "--provider", "openai"}, args...)...)
    cmd.Env = append(os.Environ(), "OPENAI_API_KEY=test", "OPENAI_API_BASE="+serverURL)
    out, err := cmd.CombinedOutput()
    return string(out), err
}

func TestInlineQuestion(t *testing.T) {
    srv := mockOpenAI()
    defer srv.Close()
    out, err := runCLI(t, srv.URL, "What is 2+2?")
    if err != nil {
        t.Fatalf("runCLI: %v\n%s", err, out)
    }
    if !strings.Contains(out, "4") {
        t.Fatalf("expected answer, got %q", out)
    }
}

func TestInlineShellSuggestion(t *testing.T) {
    srv := mockOpenAI()
    defer srv.Close()
    cmd := exec.Command("go", "run", "./cmd/cli", "--provider", "openai", "--shell", "list files")
    cmd.Env = append(os.Environ(), "OPENAI_API_KEY=test", "OPENAI_API_BASE="+srv.URL, "TERM=dumb")
    stdin, err := cmd.StdinPipe()
    if err != nil { t.Fatal(err) }
    var buf bytes.Buffer
    cmd.Stdout = &buf
    cmd.Stderr = &buf
    if err := cmd.Start(); err != nil { t.Fatal(err) }
    io.WriteString(stdin, "a\n")
    stdin.Close()
    if err := cmd.Wait(); err != nil { t.Fatalf("cmd: %v\n%s", err, buf.String()) }
    out := buf.String()
    if !strings.Contains(out, "ls -la") {
        t.Fatalf("expected command, got %q", out)
    }
}

func TestInlineFollowUp(t *testing.T) {
    srv := mockOpenAI()
    defer srv.Close()
    if _, err := runCLI(t, srv.URL, "Hello"); err != nil {
        t.Fatalf("first: %v", err)
    }
    out, err := runCLI(t, srv.URL, "Who am I?")
    if err != nil {
        t.Fatalf("second: %v\n%s", err, out)
    }
    if !strings.Contains(strings.ToLower(out), "hello") {
        t.Fatalf("expected context, got %q", out)
    }
}

func TestTUIChatMultiTurn(t *testing.T) {
    t.Skip("requires interactive terminal for full TUI interaction")
}

func TestTUIShellThenQuestion(t *testing.T) {
    t.Skip("requires interactive terminal for shell interaction")
}

func TestPipeInput(t *testing.T) {
    srv := mockOpenAI()
    defer srv.Close()
    cmd := exec.Command("bash", "-c", "echo 'What is 2+2?' | go run ./cmd/cli --provider openai")
    cmd.Env = append(os.Environ(), "OPENAI_API_KEY=test", "OPENAI_API_BASE="+srv.URL)
    out, err := cmd.CombinedOutput()
    if err != nil { t.Fatalf("pipe input: %v\n%s", err, out) }
    if !strings.Contains(string(out), "4") {
        t.Fatalf("expected answer, got %q", out)
    }
}

func TestPipeOutput(t *testing.T) {
    srv := mockOpenAI()
    defer srv.Close()
    cmd := exec.Command("bash", "-c", "go run ./cmd/cli --provider openai 'What is 2+2?' | grep 4")
    cmd.Env = append(os.Environ(), "OPENAI_API_KEY=test", "OPENAI_API_BASE="+srv.URL)
    out, err := cmd.CombinedOutput()
    if err != nil { t.Fatalf("pipe output: %v\n%s", err, out) }
    if !strings.Contains(string(out), "4") {
        t.Fatalf("expected answer, got %q", out)
    }
}

