package providers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OpenAIAdapter is a minimal adapter that performs a plain HTTP POST to an
// OpenAI-compatible endpoint and returns the response Body (io.ReadCloser).
// It intentionally avoids introducing a dependency on a specific OpenAI Go
// SDK so the repository stays light-weight; callers should provide the API
// key and base URL.
type OpenAIAdapter struct {
	ctx    context.Context
	apiKey string
	base   string
	client *http.Client
}

// NewOpenAIAdapter constructs an adapter. baseURL should be the provider base
// (for example "https://api.openai.com" or a gateway URL). If baseURL is
// empty, the adapter will default to "https://api.openai.com".
func NewOpenAIAdapter(ctx context.Context, apiKey, baseURL string) *OpenAIAdapter {
	base := strings.TrimRight(baseURL, "/")
	if base == "" {
		base = "https://api.openai.com"
	}
	return &OpenAIAdapter{ctx: ctx, apiKey: apiKey, base: base, client: &http.Client{}}
}

// Stream builds a minimal chat completion request and returns the response
// body as an io.ReadCloser for streaming consumption. The caller must close
// the returned ReadCloser when done.
func (a *OpenAIAdapter) Stream(prompt string) (io.ReadCloser, error) {
	if a == nil {
		return nil, fmt.Errorf("adapter is nil")
	}
	reqBody := strings.NewReader(`{"model":"gpt-4o-mini","messages":[{"role":"user","content":"` + prompt + `"}],"stream":true}`)
	req, err := http.NewRequestWithContext(a.ctx, http.MethodPost, a.base+"/v1/chat/completions", reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		// read small error body
		buf := make([]byte, 512)
		n, _ := resp.Body.Read(buf)
		resp.Body.Close()
		return nil, fmt.Errorf("provider error %d: %s", resp.StatusCode, string(buf[:n]))
	}
	return resp.Body, nil
}
