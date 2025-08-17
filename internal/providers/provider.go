package providers

import (
	"context"
	"io"
)

// StreamReader is a minimal interface for provider adapters that can return
// an io.ReadCloser for streaming responses.
type StreamReader interface {
	// Stream makes a request with the given prompt and returns an io.ReadCloser
	// that yields raw chunks as the provider streams. The caller is
	// responsible for closing the returned ReadCloser.
	Stream(prompt string) (io.ReadCloser, error)
}

// SyncCompleter is an optional interface adapters may implement to perform
// a non-streaming completion (single response) and return the full text.
type SyncCompleter interface {
	Complete(prompt string) (string, error)
}

// AdapterType indicates which underlying client to use.
type AdapterType string

// AdapterSDK is the AdapterType value for using the (unofficial) OpenAI SDK-based provider.
const (
	AdapterSDK  AdapterType = "sdk"
	AdapterHTTP AdapterType = "http"
)

// NewProviderAdapter returns a StreamReader implementation.
// If adapter is empty, AdapterSDK is used by default.
func NewProviderAdapter(ctx context.Context, adapter AdapterType, apiKey, baseURL, model string) StreamReader {
	if adapter == "" || adapter == AdapterSDK {
		return NewGoOpenAIAdapter(ctx, apiKey, baseURL, model)
	}
	return NewOpenAIAdapter(ctx, apiKey, baseURL)
}
