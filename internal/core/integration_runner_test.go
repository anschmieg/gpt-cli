package core

import (
	"context"
	"strings"
	"testing"

	"github.com/anschmieg/gpt-cli/internal/providers"
	"github.com/anschmieg/gpt-cli/internal/testhelpers"
)

func TestIntegration_RunStreaming_HTTPAdapter(t *testing.T) {
	chunks := []string{"hello", " world"}
	srv := testhelpers.NewChunkedServer(chunks, "application/octet-stream", 5)
	defer srv.Close()

	adapter := providers.NewOpenAIAdapter(context.Background(), "test", srv.URL)
	fr, errs, cancel := RunStreaming(context.Background(), adapter, "prompt")
	defer cancel()

	var out strings.Builder
	for s := range fr {
		out.WriteString(s)
	}
	select {
	case e := <-errs:
		if e != nil {
			t.Fatalf("error from runner: %v", e)
		}
	default:
	}

	if out.String() != "hello world" {
		t.Fatalf("unexpected combined output: %q", out.String())
	}
}

func TestIntegration_RunStreaming_SDKAdapter(t *testing.T) {
	// emulate SSE-style JSON events with data: payload
	chunks := []string{
		"data: {\"choices\":[{\"delta\":{\"content\":\"sdk\"}}]}\n\n",
		"data: {\"choices\":[{\"delta\":{\"content\":\"-done\"}}]}\n\n",
	}
	srv := testhelpers.NewChunkedServer(chunks, "text/event-stream", 5)
	defer srv.Close()

	adapter := providers.NewGoOpenAIAdapter(context.Background(), "", srv.URL, "test-model")
	fr, errs, cancel := RunStreaming(context.Background(), adapter, "prompt")
	defer cancel()

	var out strings.Builder
	for s := range fr {
		out.WriteString(s)
	}
	select {
	case e := <-errs:
		if e != nil {
			t.Fatalf("error from runner: %v", e)
		}
	default:
	}

	if out.String() != "sdk-done" {
		t.Fatalf("unexpected SDK combined output: %q", out.String())
	}
}
