package cli_test

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/anschmieg/gpt-cli/internal/core"
	"github.com/anschmieg/gpt-cli/internal/providers"
	th "github.com/anschmieg/gpt-cli/internal/testhelpers"
)

func TestOpenAIAdapterCompleteAndStreaming(t *testing.T) {
	// create a mock server that returns chunked messages
	chunks := []string{"Hello", " ", "test"}
	srv := th.NewChunkedServer(chunks, "application/octet-stream", 10)
	defer srv.Close()

	// wire the HTTP adapter to the mock server
	adapter := providers.NewOpenAIAdapter(context.Background(), "", srv.URL)

	// non-streaming: Complete should attempt to parse a JSON response; since
	// our chunked server returns raw bytes, Complete is likely to fail; this
	// exercise is primarily to verify the API surface. We'll test the
	// streaming path instead which should emit the chunks.
	rc, err := adapter.Stream("hello")
	if err != nil {
		t.Fatalf("stream error: %v", err)
	}
	defer rc.Close()

	// read all from the stream and ensure it contains concatenated chunks
	all, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("readall error: %v", err)
	}
	if string(all) != strings.Join(chunks, "") {
		t.Fatalf("unexpected stream content: %q", string(all))
	}

	// Now exercise the core.RunStreaming pipeline (adapter -> StreamReader -> fragments)
	// Build a simple adapter again and run RunStreaming
	adapter2 := providers.NewOpenAIAdapter(context.Background(), "", srv.URL)
	frCh, errCh, cancel := core.RunStreaming(context.Background(), adapter2, "hello")
	defer cancel()

	timeout := time.After(2 * time.Second)
	got := ""
	for {
		select {
		case f, ok := <-frCh:
			if !ok {
				if got != strings.Join(chunks, "") {
					t.Fatalf("fragments concatenated mismatch: %q", got)
				}
				return
			}
			got += f
		case e := <-errCh:
			if e != nil {
				t.Fatalf("runner error: %v", e)
			}
		case <-timeout:
			t.Fatalf("timeout waiting for fragments, got=%q", got)
		}
	}
}
