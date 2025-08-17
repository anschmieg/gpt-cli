package core

import (
	"os"
	"testing"
)

func TestBuildProviderOptions(t *testing.T) {
	// Ensure env is clean
	_ = setEnv("OPENAI_API_KEY", "")
	_ = setEnv("OPENAI_API_BASE", "")
	_ = setEnv("COPILOT_API_KEY", "")
	_ = setEnv("COPILOT_API_BASE", "")
	_ = setEnv("GEMINI_API_KEY", "")
	_ = setEnv("GEMINI_API_BASE", "")

	// No env vars set
	opts := BuildProviderOptions("openai")
	if opts.APIKey != "" || opts.BaseURL != "" {
		t.Fatalf("expected empty options, got %+v", opts)
	}

	// Set OpenAI env vars
	_ = setEnv("OPENAI_API_KEY", "abc123")
	_ = setEnv("OPENAI_API_BASE", "https://api.openai.test")
	opts = BuildProviderOptions("openai")
	if opts.APIKey != "abc123" || opts.BaseURL != "https://api.openai.test" {
		t.Fatalf("unexpected openai opts: %+v", opts)
	}
}

func TestBufferManagerSimpleLines(t *testing.T) {
	bm := NewBufferManager()
	frags := bm.AddChunk("hello world\nthis is a test\npartial")
	if len(frags) != 2 {
		t.Fatalf("expected 2 fragments, got %d: %+v", len(frags), frags)
	}
	if frags[0] != "hello world\n" {
		t.Fatalf("unexpected frag0: %q", frags[0])
	}
	if frags[1] != "this is a test\n" {
		t.Fatalf("unexpected frag1: %q", frags[1])
	}
	// buffer should contain 'partial'
	if bm.String() != "partial" {
		t.Fatalf("expected 'partial' in buffer, got %q", bm.String())
	}
}

func TestBufferManagerFencedCode(t *testing.T) {
	bm := NewBufferManager()
	frags := bm.AddChunk("prelude\n```") // opener only
	if len(frags) != 1 {
		t.Fatalf("expected 1 fragment for prelude, got %d", len(frags))
	}
	if frags[0] != "prelude\n" {
		t.Fatalf("unexpected prelude frag: %q", frags[0])
	}
	// now add code block content without closing
	frags = bm.AddChunk("code line 1\ncode line 2\n")
	if len(frags) != 0 {
		t.Fatalf("expected 0 fragments while inside fence, got %d", len(frags))
	}
	// now close fence
	frags = bm.AddChunk("```\nafter\n")
	if len(frags) < 1 {
		t.Fatalf("expected at least 1 fragment after closing fence, got %d", len(frags))
	}
	// the first fragment should contain the fence
	if frags[0] == "" {
		t.Fatalf("expected fence fragment, got empty")
	}
}

// setEnv helper returns the previous value
func setEnv(k, v string) string {
	prev := os.Getenv(k)
	_ = os.Setenv(k, v)
	return prev
}
