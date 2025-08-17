package testhelpers

import (
	"net/http"
	"net/http/httptest"
	"time"
)

// NewChunkedServer returns an httptest.Server that writes the provided
// chunks with small delays, flushes after each chunk, and then closes.
func NewChunkedServer(chunks []string, contentType string, delayMs int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		fl, _ := w.(http.Flusher)
		for _, c := range chunks {
			_, _ = w.Write([]byte(c))
			if fl != nil {
				fl.Flush()
			}
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}
	}))
}
