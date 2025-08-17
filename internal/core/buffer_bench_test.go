package core

import (
	"strings"
	"testing"
)

func BenchmarkBufferManager_AddChunk_Small(b *testing.B) {
	bm := NewBufferManager()
	chunk := "line one\nline two\n"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bm.AddChunk(chunk)
	}
}

func TestBufferManager_LargeStream(t *testing.T) {
	bm := NewBufferManager()
	// simulate a large stream by emitting many small chunks
	var sb strings.Builder
	for i := 0; i < 10000; i++ {
		sb.WriteString("line\n")
	}
	input := sb.String()
	// feed in small chunks of 8 bytes
	for i := 0; i < len(input); i += 8 {
		end := i + 8
		if end > len(input) {
			end = len(input)
		}
		frags := bm.AddChunk(input[i:end])
		_ = frags
	}
	// force flush remaining
	rem := bm.ForceFlush()
	if rem != "" {
		t.Fatalf("expected empty remainder after consuming large stream, got len=%d", len(rem))
	}
}
