package providers

import (
	"context"
	"fmt"
	"io"

	openai "github.com/sashabaranov/go-openai"
)

// GoOpenAIAdapter uses the sashabaranov/go-openai client to stream responses.
type GoOpenAIAdapter struct {
	ctx    context.Context
	client *openai.Client
	model  string
}

// NewGoOpenAIAdapter constructs a new adapter backed by the go-openai client.
func NewGoOpenAIAdapter(ctx context.Context, apiKey, baseURL, model string) *GoOpenAIAdapter {
	cfg := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}
	c := openai.NewClientWithConfig(cfg)
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &GoOpenAIAdapter{ctx: ctx, client: c, model: model}
}

// Stream implements StreamReader by starting a streaming chat completion and
// returning an io.ReadCloser that yields the streamed content as bytes.
func (a *GoOpenAIAdapter) Stream(prompt string) (io.ReadCloser, error) {
	if a == nil || a.client == nil {
		return nil, fmt.Errorf("adapter not initialized")
	}

	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		req := openai.ChatCompletionRequest{
			Model:    a.model,
			Messages: []openai.ChatCompletionMessage{{Role: "user", Content: prompt}},
			Stream:   true,
		}
		stream, err := a.client.CreateChatCompletionStream(a.ctx, req)
		if err != nil {
			pw.CloseWithError(err)
			return
		}
		defer stream.Close()

		for {
			evt, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					return
				}
				pw.CloseWithError(err)
				return
			}
			// write the content chunk to the pipe; evt is a struct so check fields
			if len(evt.Choices) > 0 {
				delta := evt.Choices[0].Delta
				if delta.Content != "" {
					_, _ = pw.Write([]byte(delta.Content))
				}
			}
		}
	}()

	return pr, nil
}

// Complete performs a single-shot chat completion using the SDK client and
// returns the assistant content. This makes the adapter usable in
// non-streaming mode when callers prefer a synchronous response.
func (a *GoOpenAIAdapter) Complete(prompt string) (string, error) {
	if a == nil || a.client == nil {
		return "", fmt.Errorf("adapter not initialized")
	}
	req := openai.ChatCompletionRequest{
		Model:    a.model,
		Messages: []openai.ChatCompletionMessage{{Role: "user", Content: prompt}},
		Stream:   false,
	}
	resp, err := a.client.CreateChatCompletion(a.ctx, req)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", nil
	}
	return resp.Choices[0].Message.Content, nil
}
