package cli_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	th "github.com/anschmieg/gpt-cli/internal/testhelpers"
)

// TestCLIEndToEnd runs the CLI (built binary) against the mock-openai
// HTTP mock server and asserts the streamed output matches the expected
// concatenated chunks.
func TestCLIEndToEnd(t *testing.T) {
	chunks := []string{"Hello", " ", "e2e"}
	srv := th.NewChunkedServer(chunks, "application/octet-stream", 10)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// build the CLI binary into a temp dir for a stable run
	td := t.TempDir()
	bin := filepath.Join(td, "gpt-cli-bin")

	// NOTE: tests run with working dir = package dir (internal/cli), so the
	// cmd path is relative to that: ../../cmd/gpt-cli
	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", bin, "../../cmd/gpt-cli")
	buildCmd.Env = os.Environ()
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\noutput:\n%s", err, string(out))
	}

	// run the built binary and capture combined output
	runCmd := exec.CommandContext(ctx, bin, "--stream", "--provider", "http", "--base-url", srv.URL, "ping")
	runCmd.Env = os.Environ()
	out, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("cmd failed: %v\noutput:\n%s", err, string(out))
	}

	got := strings.TrimSpace(string(out))
	want := strings.Join(chunks, "")
	if got != want {
		t.Fatalf("unexpected e2e output: got=%q want=%q\nfull output:\n%s", got, want, string(out))
	}
}
