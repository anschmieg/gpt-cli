package core

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"
)

// testAdapter returns an io.ReadCloser from provided chunks
type testAdapter struct {
	chunks []string
}

func (t *testAdapter) Stream(prompt string) (io.ReadCloser, error) {
	return io.NopCloser(&chunkReader{chunks: t.chunks}), nil
}

func TestRunStreaming_HappyPath(t *testing.T) {
	ta := &testAdapter{chunks: []string{"hello\n", "world\n"}}
	fr, errs, cancel := RunStreaming(context.Background(), ta, "prompt")
	defer cancel()

	var out strings.Builder
	for s := range fr {
		out.WriteString(s)
	}
	select {
	case e := <-errs:
		if e != nil {
			t.Fatalf("unexpected error: %v", e)
		}
	default:
	}

	if out.String() != "hello\nworld\n" {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestRunStreaming_FencedBlock(t *testing.T) {
	// fence opener split across chunks; should not emit fenced content until close
	chunks := []string{"pre\n```go\ncode line\n", "more code\n``\npost\n"}
	ta := &testAdapter{chunks: chunks}
	fr, errs, cancel := RunStreaming(context.Background(), ta, "prompt")
	defer cancel()

	var out strings.Builder
	for s := range fr {
		out.WriteString(s)
	}
	select {
	case e := <-errs:
		if e != nil {
			t.Fatalf("err: %v", e)
		}
	default:
	}

	got := out.String()
	if !strings.Contains(got, "pre") || !strings.Contains(got, "post") {
		t.Fatalf("missing fragments: %q", got)
	}
}

func TestRunStreaming_Cancel(t *testing.T) {
	// long-running chunks; cancel after a short delay
	chunks := []string{"line1\n", "line2\n", "line3\n"}
	ta := &testAdapter{chunks: chunks}
	ctx, cancelAll := context.WithCancel(context.Background())
	fr, errs, cancel := RunStreaming(ctx, ta, "prompt")
	// cancel the inner cancel func after a bit
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancelAll()
		cancel()
	}()

	// drain fragments until closed
	var count int
	for range fr {
		count++
	}
	select {
	case e := <-errs:
		if e != nil {
			t.Fatalf("unexpected err: %v", e)
		}
	default:
	}
	if count == 0 {
		t.Fatalf("expected at least one fragment before cancel")
	}
}
