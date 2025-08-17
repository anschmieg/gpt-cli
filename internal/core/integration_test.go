package core

import (
	"fmt"
	"strings"
	"testing"
)

func TestIntegration_StreamReaderWithBufferManager(t *testing.T) {
	// Test the full integration of StreamReader + BufferManager
	// with a realistic streaming scenario

	input := "This is a test.\n\nHere is some code:\n```go\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello, world!\")\n}\n```\n\nAnd some more text."

	// Simulate chunked streaming (as would happen with a real provider)
	chunks := []string{
		"This is a test.\n\nHere is some code:\n",
		"```go\npackage main\n\nimport \"fmt\"\n\n",
		"func main() {\n\tfmt.Println(\"Hello, world!\")\n}\n",
		"```\n\nAnd some more text.",
	}

	// Create a reader that yields chunks
	reader := &chunkReader{chunks: chunks}

	// Use StreamReader to process the stream
	ch, err := StreamReader(reader)
	if err != nil {
		t.Fatalf("StreamReader error: %v", err)
	}

	var result strings.Builder
	var frags []string
	for s := range ch {
		frags = append(frags, s)
		result.WriteString(s)
	}

	expected := input
	got := result.String()
	if got != expected {
		t.Errorf("integration test failed:\nExpected: %q\nGot: %q\nFragments: %+v", expected, got, frags)
	}
}

func TestIntegration_BufferManagerEdgeCases(t *testing.T) {
	// Test edge cases that might cause issues with buffer management

	testCases := []struct {
		n string
		i []string
		e string
	}{
		{
			n: "partial fence at end",
			i: []string{"Here is code:\n```go\npackage main", "\n```"},
			e: "Here is code:\n```go\npackage main\n```",
		},
		{
			n: "multiple fences",
			i: []string{"First:\n```\ncode1\n```\nSecond:\n```\ncode2\n```"},
			e: "First:\n```\ncode1\n```\nSecond:\n```\ncode2\n```",
		},
		{
			n: "nested fences (should treat as separate)",
			i: []string{"Outer:\n```\nInner:\n```\ncode\n```\n```"},
			e: "Outer:\n```\nInner:\n```\ncode\n```\n```",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.n, func(t *testing.T) {
			reader := &chunkReader{chunks: tc.i}
			ch, err := StreamReader(reader)
			if err != nil {
				t.Fatalf("StreamReader error: %v", err)
			}

			var result strings.Builder
			for s := range ch {
				result.WriteString(s)
			}

			got := result.String()
			if got != tc.e {
				t.Errorf("case %q failed:\nExpected: %q\nGot: %q", tc.n, tc.e, got)
			}
		})
	}
}

func TestIntegration_LongStreamHandling(t *testing.T) {
	// Test handling of very long streams to ensure no memory issues
	var longInput strings.Builder
	longInput.WriteString("Start of long content\n")

	// Add 1000 lines of content
	for i := 0; i < 1000; i++ {
		longInput.WriteString(fmt.Sprintf("Line %d: This is some test content that should be handled efficiently.\n", i))
	}

	longInput.WriteString("End of long content.")

	// Split into chunks. Re-append newlines except for the last empty split
	parts := strings.Split(longInput.String(), "\n")
	chunks := make([]string, len(parts))
	for i := range parts {
		if i < len(parts)-1 {
			chunks[i] = parts[i] + "\n"
		} else {
			chunks[i] = parts[i]
		}
	}

	reader := &chunkReader{chunks: chunks}
	ch, err := StreamReader(reader)
	if err != nil {
		t.Fatalf("StreamReader error: %v", err)
	}

	var result strings.Builder
	for s := range ch {
		result.WriteString(s)
	}

	expected := longInput.String()
	got := result.String()
	if got != expected {
		t.Errorf("long stream test failed - lengths differ: expected %d, got %d", len(expected), len(got))
	}
}

// Test the full provider options workflow
func TestIntegration_ProviderOptionsWorkflow(t *testing.T) {
	// Test the full workflow from env vars to provider options
	origOpenAIKey := setEnv("OPENAI_API_KEY", "test-openai-key")
	origOpenAIBase := setEnv("OPENAI_API_BASE", "https://api.openai.test")
	origCopilotKey := setEnv("COPILOT_API_KEY", "test-copilot-key")
	origCopilotBase := setEnv("COPILOT_API_BASE", "https://api.copilot.test")
	origGeminiKey := setEnv("GEMINI_API_KEY", "test-gemini-key")
	origGeminiBase := setEnv("GEMINI_API_BASE", "https://api.gemini.test")

	defer func() {
		setEnv("OPENAI_API_KEY", origOpenAIKey)
		setEnv("OPENAI_API_BASE", origOpenAIBase)
		setEnv("COPILOT_API_KEY", origCopilotKey)
		setEnv("COPILOT_API_BASE", origCopilotBase)
		setEnv("GEMINI_API_KEY", origGeminiKey)
		setEnv("GEMINI_API_BASE", origGeminiBase)
	}()

	providers := []string{"openai", "copilot", "gemini", "unknown"}
	expected := map[string]*ProviderOptions{
		"openai":  {APIKey: "test-openai-key", BaseURL: "https://api.openai.test"},
		"copilot": {APIKey: "test-copilot-key", BaseURL: "https://api.copilot.test"},
		"gemini":  {APIKey: "test-gemini-key", BaseURL: "https://api.gemini.test"},
		"unknown": {APIKey: "", BaseURL: ""},
	}

	for _, provider := range providers {
		opts := BuildProviderOptions(provider)
		exp, ok := expected[provider]
		if !ok {
			t.Fatalf("unknown provider in test: %s", provider)
		}

		if opts.APIKey != exp.APIKey || opts.BaseURL != exp.BaseURL {
			t.Errorf("provider %s: expected %+v, got %+v", provider, exp, opts)
		}

		// Test validation
		if provider != "unknown" {
			err := opts.Validate()
			if err != nil {
				t.Errorf("provider %s: unexpected validation error: %v", provider, err)
			}
		} else {
			err := opts.Validate()
			if err == nil {
				t.Errorf("provider %s: expected validation error for empty API key", provider)
			}
		}
	}
}

// setEnv helper is defined in another test file; reuse that implementation.
