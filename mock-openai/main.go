package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// minimal request shape
type chatReq struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
	Stream bool `json:"stream"`
}

func streamSSE(w http.ResponseWriter, messages []string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	// emit incremental chat completion chunk events roughly compatible with
	// OpenAI's streaming shape. Each event contains minimal metadata and a
	// choices array with a delta object containing content.
	for i, m := range messages {
		payload := fmt.Sprintf(
			`{"id":"mock-%d","object":"chat.completion.chunk","model":"mock-model","choices":[{"delta":{"content":"%s"},"index":%d}]}`,
			time.Now().UnixNano(), m, i,
		)
		fmt.Fprintf(w, "data: %s\n\n", payload)
		fl.Flush()
		time.Sleep(50 * time.Millisecond)
	}

	// final sentinel as per OpenAI: a single line with [DONE]
	fmt.Fprint(w, "data: [DONE]\n\n")
	fl.Flush()
}

func streamChunked(w http.ResponseWriter, messages []string) {
	w.Header().Set("Content-Type", "application/octet-stream")
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	for _, m := range messages {
		fmt.Fprint(w, m)
		fl.Flush()
		time.Sleep(50 * time.Millisecond)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	var req chatReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	// decide streaming style by query param (style=chunked) or the request's
	// stream boolean. If stream is false, return a full JSON response that
	// resembles the OpenAI chat completion response.
	msgs := []string{"hello", " ", "world"}
	if !req.Stream {
		// build a simple non-streaming response
		type choice struct {
			Index   int `json:"index"`
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		}
		resp := struct {
			ID      string   `json:"id"`
			Object  string   `json:"object"`
			Created int64    `json:"created"`
			Model   string   `json:"model"`
			Choices []choice `json:"choices"`
		}{
			ID:      fmt.Sprintf("mock-%d", time.Now().UnixNano()),
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   req.Model,
		}
		// assemble content into one message
		full := strings.Join(msgs, "")
		c := choice{Index: 0, FinishReason: "stop"}
		c.Message.Role = "assistant"
		c.Message.Content = full
		resp.Choices = []choice{c}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	// streaming requested: allow style=chunked to force raw chunked bytes
	style := r.URL.Query().Get("style")
	if style == "chunked" {
		streamChunked(w, msgs)
		return
	}
	streamSSE(w, msgs)
}

func main() {
	addr := flag.String("addr", ":8080", "address to listen on")
	flag.Parse()

	http.HandleFunc("/v1/chat/completions", handler)
	log.Printf("mock-openai: listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
