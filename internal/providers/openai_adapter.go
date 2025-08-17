package providers

import (
	"context"
	"encoding/json"
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

// Complete performs a single-shot (non-streaming) chat completion and returns
// the concatenated assistant message. This implements the SyncCompleter
// optional interface so callers can switch between streaming and non-streaming
// modes.
func (a *OpenAIAdapter) Complete(prompt string) (string, error) {
	if a == nil {
		return "", fmt.Errorf("adapter is nil")
	}
	reqBody := strings.NewReader(`{"model":"gpt-4o-mini","messages":[{"role":"user","content":"` + prompt + `"}],"stream":false}`)
	req, err := http.NewRequestWithContext(a.ctx, http.MethodPost, a.base+"/v1/chat/completions", reqBody)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		buf := make([]byte, 512)
		n, _ := resp.Body.Read(buf)
		return "", fmt.Errorf("provider error %d: %s", resp.StatusCode, string(buf[:n]))
	}
	// naive parsing: read body and attempt to extract the assistant content
	// (the mock and typical providers return the assistant message at
	// choices[0].message.content). For robust parsing we decode into a struct.
	type choiceMsg struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	}
	var respObj struct {
		ID      string      `json:"id"`
		Object  string      `json:"object"`
		Created int64       `json:"created"`
		Model   string      `json:"model"`
		Choices []choiceMsg `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respObj); err != nil {
		return "", err
	}
	if len(respObj.Choices) == 0 {
		return "", nil
	}
	return respObj.Choices[0].Message.Content, nil
}
