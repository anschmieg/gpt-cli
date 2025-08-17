package providers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anschmieg/gpt-cli/internal/providers"
)

func TestOpenAIAdapterCompleteParsesJSON(t *testing.T) {
	// create a mock JSON response similar to OpenAI non-streaming
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"id":      "mock-1",
			"object":  "chat.completion",
			"created": 1,
			"model":   "mock",
			"choices": []map[string]interface{}{{
				"index":         0,
				"message":       map[string]string{"role": "assistant", "content": "ok done"},
				"finish_reason": "stop",
			}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	a := providers.NewOpenAIAdapter(context.Background(), "", srv.URL)
	got, err := a.Complete("hi")
	if err != nil {
		t.Fatalf("Complete error: %v", err)
	}
	if got != "ok done" {
		t.Fatalf("unexpected complete content: %q", got)
	}
}
