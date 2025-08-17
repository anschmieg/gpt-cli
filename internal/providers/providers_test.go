package providers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOpenAIHTTPAdapter_Stream(t *testing.T) {
	// httptest server that writes two chunks and flushes them
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		fl, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("response is not flushable")
		}
		_, _ = w.Write([]byte("hello"))
		fl.Flush()
		time.Sleep(10 * time.Millisecond)
		_, _ = w.Write([]byte(" world"))
		fl.Flush()
	}))
	defer srv.Close()

	a := NewOpenAIAdapter(context.Background(), "test-key", srv.URL)
	rc, err := a.Stream("hi")
	if err != nil {
		t.Fatalf("Stream error: %v", err)
	}
	defer rc.Close()

	b, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	got := string(b)
	if got != "hello world" {
		t.Fatalf("unexpected body: %q", got)
	}
}

func TestGoOpenAIAdapter_Stream(t *testing.T) {
	// SSE-like stream: send two data: events then [DONE]
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fl, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("response is not flushable")
		}
		// Event 1
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"hello\"}}]}\n\n"))
		fl.Flush()
		time.Sleep(10 * time.Millisecond)
		// Event 2
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\" world\"}}]}\n\n"))
		fl.Flush()
		time.Sleep(10 * time.Millisecond)
		// Done
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
		fl.Flush()
	}))
	defer srv.Close()

	a := NewGoOpenAIAdapter(context.Background(), "", srv.URL, "test-model")
	rc, err := a.Stream("hi")
	if err != nil {
		t.Fatalf("Stream error: %v", err)
	}
	defer rc.Close()

	b, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	got := string(b)
	if got != "hello world" {
		t.Fatalf("unexpected body from SDK stream: %q", got)
	}
}
