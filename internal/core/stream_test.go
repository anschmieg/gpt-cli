package core

import (
	"io"
	"strings"
	"testing"
)

func TestStreamReaderSimpleLines(t *testing.T) {
	input := "hello world\nthis is a test\npartial"
	r := strings.NewReader(input)
	ch, err := StreamReader(r)
	if err != nil {
		t.Fatalf("StreamReader error: %v", err)
	}

	var out []string
	for s := range ch {
		out = append(out, s)
	}

	// Be tolerant to how the underlying reader splits chunks. Reconstruct the
	// emitted output and compare to the original input.
	got := strings.Join(out, "")
	if got != input {
		t.Fatalf("reconstructed output mismatch: expected %q, got %q (fragments: %+v)", input, got, out)
	}
}

func TestStreamReaderFencedCode(t *testing.T) {
	// simulate chunked streaming where fence opener and body arrive
	// in separate reads
	chunks := []string{
		"prelude\n```\ncode line 1\n",
		"code line 2\n```\nafter\n",
	}
	// Build an io.Reader that yields the chunks in sequence
	r := io.NopCloser(&chunkReader{chunks: chunks})
	ch, err := StreamReader(r)
	if err != nil {
		t.Fatalf("StreamReader error: %v", err)
	}

	var got []string
	for s := range ch {
		got = append(got, s)
	}

	// We expect prelude to be emitted, then the fenced block including fences
	if len(got) < 2 {
		t.Fatalf("expected at least 2 fragments, got %d: %+v", len(got), got)
	}
	if got[0] != "prelude\n" {
		t.Fatalf("unexpected first fragment: %q", got[0])
	}
	// the fenced fragment should contain at least one code line and closing fence
	if !strings.Contains(got[1], "code line 2") || !strings.Contains(got[1], "```") {
		t.Fatalf("unexpected fenced fragment: %q", got[1])
	}
}

// chunkReader implements io.Reader and returns each chunk on successive reads.
// It ignores the provided buffer size and returns chunk bytes until exhausted.
type chunkReader struct {
	chunks []string
	idx    int
}

func (c *chunkReader) Read(p []byte) (int, error) {
	if c.idx >= len(c.chunks) {
		return 0, io.EOF
	}
	s := c.chunks[c.idx]
	c.idx++
	n := copy(p, []byte(s))
	return n, nil
}

func (c *chunkReader) Close() error { return nil }
