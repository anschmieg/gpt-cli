//go:build integration
// +build integration

package providers

import (
    "bytes"
    "encoding/json"
    "io"
    "net/http"
    "testing"

    "github.com/anschmieg/gpt-cli/internal/config"
    "github.com/stretchr/testify/assert"
)

// Helpers build minimal JSON payloads matching provider expectations without
// anonymous struct composite literals.

func writeJSON(t *testing.T, status int, v any) *http.Response {
    t.Helper()
    var buf bytes.Buffer
    _ = json.NewEncoder(&buf).Encode(v)
    return &http.Response{
        StatusCode: status,
        Header:     http.Header{"Content-Type": []string{"application/json"}},
        Body:       io.NopCloser(bytes.NewReader(buf.Bytes())),
    }
}

type roundTripFunc func(*http.Request) *http.Response

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
    return f(r), nil
}

func TestOpenAIProvider_CallProvider_Success(t *testing.T) {
    cfg := &config.Config{Provider: "openai", Model: "gpt-4", APIKey: "test", BaseURL: "https://example.com"}
    p := NewOpenAIProvider(cfg)
    p.client.Transport = roundTripFunc(func(r *http.Request) *http.Response {
        assert.Equal(t, "/v1/chat/completions", r.URL.Path)
        return writeJSON(t, http.StatusOK, map[string]any{
            "choices": []any{
                map[string]any{"message": map[string]any{"content": "ok from openai"}},
            },
        })
    })
    got, err := p.CallProvider("hello")
    assert.NoError(t, err)
    assert.Equal(t, "ok from openai", got)
}

func TestOpenAIProvider_CallProvider_HTTPError(t *testing.T) {
    cfg := &config.Config{Provider: "openai", Model: "gpt-4", APIKey: "test", BaseURL: "https://example.com"}
    p := NewOpenAIProvider(cfg)
    p.client.Transport = roundTripFunc(func(r *http.Request) *http.Response {
        return &http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(bytes.NewBufferString("boom"))}
    })
    got, err := p.CallProvider("hello")
    assert.Error(t, err)
    assert.Empty(t, got)
}

func TestOpenAIProvider_CallProvider_ErrorField(t *testing.T) {
    cfg := &config.Config{Provider: "openai", Model: "gpt-4", APIKey: "test", BaseURL: "https://example.com"}
    p := NewOpenAIProvider(cfg)
    p.client.Transport = roundTripFunc(func(r *http.Request) *http.Response {
        return writeJSON(t, http.StatusOK, map[string]any{
            "error": map[string]any{"message": "quota exceeded", "code": "rate_limit"},
        })
    })
    _, err := p.CallProvider("hello")
    assert.Error(t, err)
}

func TestOpenAIProvider_StreamProvider_Chunks(t *testing.T) {
    cfg := &config.Config{Provider: "openai", Model: "gpt-4", APIKey: "test", BaseURL: "https://example.com"}
    p := NewOpenAIProvider(cfg)
    p.client.Transport = roundTripFunc(func(r *http.Request) *http.Response {
        return writeJSON(t, http.StatusOK, map[string]any{
            "choices": []any{
                map[string]any{"message": map[string]any{"content": "Hello world"}},
            },
        })
    })
    cc, ec := p.StreamProvider("hi")
    var chunks []string
    for s := range cc {
        chunks = append(chunks, s)
    }
    // error channel should be closed with no error
    if err, ok := <-ec; ok {
        assert.NoError(t, err)
    }
    // Simulated streaming splits on words with trailing spaces
    assert.Equal(t, []string{"Hello ", "world "}, chunks)
}

func TestOpenAIProvider_StreamProvider_Error(t *testing.T) {
    cfg := &config.Config{Provider: "openai", Model: "gpt-4", APIKey: "test", BaseURL: "https://example.com"}
    p := NewOpenAIProvider(cfg)
    p.client.Transport = roundTripFunc(func(r *http.Request) *http.Response {
        return &http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(bytes.NewBufferString("nope"))}
    })
    cc, ec := p.StreamProvider("hi")
    // content should close immediately; errorChan should yield an error
    for range cc {
        // drain
    }
    if err, ok := <-ec; ok {
        assert.Error(t, err)
    } else {
        t.Fatalf("expected error on errorChan")
    }
}

func TestCopilotProvider_URLVariants(t *testing.T) {
    // Accept any path but record what we got
    var gotPath string

    t.Run("base without v1", func(t *testing.T) {
        cfg := &config.Config{Provider: "copilot", Model: "m", APIKey: "k", BaseURL: "https://example.com"}
        p := NewCopilotProvider(cfg)
        p.client.Transport = roundTripFunc(func(r *http.Request) *http.Response {
            gotPath = r.URL.Path
            return writeJSON(t, http.StatusOK, map[string]any{"choices": []any{map[string]any{"message": map[string]any{"content": "ok"}}}})
        })
        _, err := p.CallProvider("hi")
        assert.NoError(t, err)
        assert.Equal(t, "/v1/chat/completions", gotPath)
    })

    t.Run("base with /v1", func(t *testing.T) {
        cfg := &config.Config{Provider: "copilot", Model: "m", APIKey: "k", BaseURL: "https://example.com/v1"}
        p := NewCopilotProvider(cfg)
        p.client.Transport = roundTripFunc(func(r *http.Request) *http.Response {
            gotPath = r.URL.Path
            return writeJSON(t, http.StatusOK, map[string]any{"choices": []any{map[string]any{"message": map[string]any{"content": "ok"}}}})
        })
        _, err := p.CallProvider("hi")
        assert.NoError(t, err)
        assert.Equal(t, "/v1/chat/completions", gotPath)
    })

    t.Run("already full path", func(t *testing.T) {
        cfg := &config.Config{Provider: "copilot", Model: "m", APIKey: "k", BaseURL: "https://example.com/v1/chat/completions"}
        p := NewCopilotProvider(cfg)
        p.client.Transport = roundTripFunc(func(r *http.Request) *http.Response {
            gotPath = r.URL.Path
            return writeJSON(t, http.StatusOK, map[string]any{"choices": []any{map[string]any{"message": map[string]any{"content": "ok"}}}})
        })
        _, err := p.CallProvider("hi")
        assert.NoError(t, err)
        assert.Equal(t, "/v1/chat/completions", gotPath)
    })
}

func TestGeminiProvider_CallProvider_Success(t *testing.T) {
    cfg := &config.Config{Provider: "gemini", Model: "m", APIKey: "k", BaseURL: "https://example.com"}
    p := NewGeminiProvider(cfg)
    p.client.Transport = roundTripFunc(func(r *http.Request) *http.Response {
        assert.Equal(t, "/chat/completions", r.URL.Path)
        return writeJSON(t, http.StatusOK, map[string]any{"choices": []any{map[string]any{"message": map[string]any{"content": "ok from gemini"}}}})
    })
    got, err := p.CallProvider("hi")
    assert.NoError(t, err)
    assert.Equal(t, "ok from gemini", got)
}

func TestCopilotProvider_HTTPErrorAndMalformedJSON(t *testing.T) {
    cfg := &config.Config{Provider: "copilot", Model: "m", APIKey: "k", BaseURL: "https://example.com"}
    p := NewCopilotProvider(cfg)
    // HTTP error
    p.client.Transport = roundTripFunc(func(r *http.Request) *http.Response {
        return &http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(bytes.NewBufferString("bad"))}
    })
    _, err := p.CallProvider("hi")
    assert.Error(t, err)

    // Malformed JSON
    p.client.Transport = roundTripFunc(func(r *http.Request) *http.Response {
        return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString("{"))}
    })
    _, err = p.CallProvider("hi")
    assert.Error(t, err)
}

func TestGeminiProvider_ErrorFieldAndMalformedJSON(t *testing.T) {
    cfg := &config.Config{Provider: "gemini", Model: "m", APIKey: "k", BaseURL: "https://example.com"}
    p := NewGeminiProvider(cfg)
    // Error field
    p.client.Transport = roundTripFunc(func(r *http.Request) *http.Response {
        return writeJSON(t, http.StatusOK, map[string]any{"error": map[string]any{"message": "err", "code": "c"}})
    })
    _, err := p.CallProvider("hi")
    assert.Error(t, err)

    // Malformed JSON
    p.client.Transport = roundTripFunc(func(r *http.Request) *http.Response {
        return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString("{"))}
    })
    _, err = p.CallProvider("hi")
    assert.Error(t, err)
}

func TestOpenAIProvider_CallProvider_NoChoices(t *testing.T) {
    cfg := &config.Config{Provider: "openai", Model: "m", APIKey: "k", BaseURL: "https://example.com"}
    p := NewOpenAIProvider(cfg)
    p.client.Transport = roundTripFunc(func(r *http.Request) *http.Response {
        return writeJSON(t, http.StatusOK, map[string]any{"choices": []any{}})
    })
    out, err := p.CallProvider("hi")
    assert.Error(t, err)
    assert.Empty(t, out)
}

func TestOpenAIProvider_CallProvider_MalformedJSON(t *testing.T) {
    cfg := &config.Config{Provider: "openai", Model: "m", APIKey: "k", BaseURL: "https://example.com"}
    p := NewOpenAIProvider(cfg)
    p.client.Transport = roundTripFunc(func(r *http.Request) *http.Response {
        return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString("{"))}
    })
    _, err := p.CallProvider("hi")
    assert.Error(t, err)
}

func TestCopilotProvider_StreamProvider_SplitsWords(t *testing.T) {
    cfg := &config.Config{Provider: "copilot", Model: "m", APIKey: "k", BaseURL: "https://example.com"}
    p := NewCopilotProvider(cfg)
    p.client.Transport = roundTripFunc(func(r *http.Request) *http.Response {
        return writeJSON(t, http.StatusOK, map[string]any{"choices": []any{map[string]any{"message": map[string]any{"content": "foo bar"}}}})
    })
    cc, ec := p.StreamProvider("hi")
    var chunks []string
    for s := range cc { chunks = append(chunks, s) }
    if err, ok := <-ec; ok { assert.NoError(t, err) }
    assert.Equal(t, []string{"foo ", "bar "}, chunks)
}

func TestGeminiProvider_StreamProvider_SplitsWords(t *testing.T) {
    cfg := &config.Config{Provider: "gemini", Model: "m", APIKey: "k", BaseURL: "https://example.com"}
    p := NewGeminiProvider(cfg)
    p.client.Transport = roundTripFunc(func(r *http.Request) *http.Response {
        return writeJSON(t, http.StatusOK, map[string]any{"choices": []any{map[string]any{"message": map[string]any{"content": "lorem ipsum"}}}})
    })
    cc, ec := p.StreamProvider("hi")
    var chunks []string
    for s := range cc { chunks = append(chunks, s) }
    if err, ok := <-ec; ok { assert.NoError(t, err) }
    assert.Equal(t, []string{"lorem ", "ipsum "}, chunks)
}
