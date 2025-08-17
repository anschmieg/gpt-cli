package providers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIntegration_Adapters(t *testing.T) {
	// HTTP server for the plain HTTP adapter
	httpSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		fl, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("response is not flushable")
		}
		_, _ = w.Write([]byte("adapter-http"))
		fl.Flush()
		time.Sleep(5 * time.Millisecond)
		_, _ = w.Write([]byte("-done"))
		fl.Flush()
	}))
	defer httpSrv.Close()

	// HTTP server for the SDK adapter; emulate SSE-style events
	sdkSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fl, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("response is not flushable")
		}
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"adapter-sdk\"}}]}\n\n"))
		fl.Flush()
		time.Sleep(5 * time.Millisecond)
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"-done\"}}]}\n\n"))
		fl.Flush()
		// final done
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
		fl.Flush()
	}))
	defer sdkSrv.Close()

	ctx := context.Background()

	// Test SDK adapter selection by default
	sdkAdapter := NewProviderAdapter(ctx, AdapterSDK, "", sdkSrv.URL, "test-model")
	rc, err := sdkAdapter.Stream("hi")
	if err != nil {
		t.Fatalf("sdk adapter stream error: %v", err)
	}
	b, err := io.ReadAll(rc)
	rc.Close()
	if err != nil {
		t.Fatalf("sdk read error: %v", err)
	}
	if string(b) != "adapter-sdk-done" {
		t.Fatalf("sdk adapter unexpected: %q", string(b))
	}

	// Test HTTP adapter selection
	httpAdapter := NewProviderAdapter(ctx, AdapterHTTP, "", httpSrv.URL, "")
	rc2, err := httpAdapter.Stream("hi")
	if err != nil {
		t.Fatalf("http adapter stream error: %v", err)
	}
	b2, err := io.ReadAll(rc2)
	rc2.Close()
	if err != nil {
		t.Fatalf("http read error: %v", err)
	}
	if string(b2) != "adapter-http-done" {
		t.Fatalf("http adapter unexpected: %q", string(b2))
	}
}
