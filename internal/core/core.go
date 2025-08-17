package core

import (
	"fmt"
	"os"
	"strings"
)

// Config contains runtime options used by the core logic.
type Config struct {
	Provider    string
	Model       string
	Temperature float64
	System      string
	Prompt      string
	Stream      bool
}

// ProviderOptions holds API credentials and base URL for a provider.
type ProviderOptions struct {
	APIKey  string
	BaseURL string
}

// Validate performs lightweight validation of provider options. It returns an
// error when required fields (like APIKey) are not present. This is additive
// and callers may still choose to accept empty options for test scenarios.
func (p *ProviderOptions) Validate() error {
	if p == nil {
		return fmt.Errorf("provider options is nil")
	}
	if strings.TrimSpace(p.APIKey) == "" {
		return fmt.Errorf("missing API key")
	}
	return nil
}

// ProviderResponse contains text and markdown returned by providers.
type ProviderResponse struct {
	Text     string
	Markdown string
}

// BuildProviderOptions reads environment variables and returns ProviderOptions
// for the requested provider. It's intentionally simple and testable.
func BuildProviderOptions(provider string) *ProviderOptions {
	p := strings.ToLower(provider)
	opts := &ProviderOptions{}
	switch p {
	case "openai":
		opts.APIKey = os.Getenv("OPENAI_API_KEY")
		opts.BaseURL = os.Getenv("OPENAI_API_BASE")
	case "copilot":
		opts.APIKey = os.Getenv("COPILOT_API_KEY")
		opts.BaseURL = os.Getenv("COPILOT_API_BASE")
	case "gemini":
		opts.APIKey = os.Getenv("GEMINI_API_KEY")
		opts.BaseURL = os.Getenv("GEMINI_API_BASE")
	default:
		// Unknown provider: leave empty options.
	}
	return opts
}

// BufferManager turns arbitrary incoming chunks into "safe fragments" that
// can be rendered by a markdown renderer without producing incorrect output
// when markdown tokens cross chunk boundaries (eg. fenced code blocks).
//
// The algorithm is intentionally conservative: it flushes on newline boundaries
// and keeps fenced code blocks intact until the closing fence is observed.
type BufferManager struct {
	buf       strings.Builder
	inFence   bool
	fenceMark string // e.g. ``` or ~~~
}

// NewBufferManager creates a new BufferManager.
func NewBufferManager() *BufferManager {
	return &BufferManager{}
}

// AddChunk appends a new chunk and returns any safe fragments that can be
// rendered now. Fragments will include trailing newlines where applicable.
func (b *BufferManager) AddChunk(chunk string) []string {
	var out []string
	b.buf.WriteString(chunk)

	for {
		s := b.buf.String()
		if s == "" {
			break
		}

		if b.inFence {
			// Look for closing fence marker AFTER the opener. The buffer
			// begins with the opener we preserved, so searching from index
			// len(fenceMark) avoids matching the opener itself.
			restStart := len(b.fenceMark)
			closeIdx := strings.Index(s[restStart:], b.fenceMark)
			if closeIdx == -1 {
				// No closing fence yet: yield nothing (keep buffering)
				break
			}
			// closeIdx is relative to s[restStart:], compute absolute position
			closePos := restStart + closeIdx
			// Include the closing fence in the fragment
			end := closePos + len(b.fenceMark)
			frag := s[:end]
			out = append(out, frag)
			// Consume from buffer
			rest := s[end:]
			b.buf.Reset()
			b.buf.WriteString(rest)
			b.inFence = false
			b.fenceMark = ""
			// continue to see if more can be emitted
			continue
		}

		// Not in fence: look for the next fence opener or last newline
		fenceIdx := -1
		fenceMark := ""
		// check for ``` or ~~~
		if i := strings.Index(s, "```"); i != -1 {
			fenceIdx = i
			fenceMark = "```"
		}
		if j := strings.Index(s, "~~~"); j != -1 {
			// prefer earlier occurrence
			if fenceIdx == -1 || j < fenceIdx {
				fenceIdx = j
				fenceMark = "~~~"
			}
		}

		if fenceIdx != -1 {
			// there is a fence opener; flush up to the opener (safe)
			if fenceIdx > 0 {
				frag := s[:fenceIdx]
				out = append(out, frag)
			}
			// Now start fence: keep the opener and the remainder in the buffer
			// so that when the closing fence is seen we can emit the entire
			// fenced block including the opener (and optional language tag).
			rest := s[fenceIdx:]
			b.buf.Reset()
			b.buf.WriteString(rest)
			b.inFence = true
			b.fenceMark = fenceMark
			break
		}

		// No fence present. Flush up to the last newline if any.
		// Flush incrementally on the first newline found so that callers get
		// smaller fragments (line-by-line) instead of large combined blocks.
		if idx := strings.Index(s, "\n"); idx != -1 {
			frag := s[:idx+1]
			out = append(out, frag)
			rest := s[idx+1:]
			b.buf.Reset()
			b.buf.WriteString(rest)
			// continue to see if more can be emitted
			continue
		}

		// No newline and no fence: don't emit yet (avoid mid-token flush)
		break
	}

	return out
}

// ForceFlush drains the buffer and returns whatever remains. Use carefully in
// situations where you must emit partial content (with best-effort correctness).
func (b *BufferManager) ForceFlush() string {
	s := b.buf.String()
	b.buf.Reset()
	b.inFence = false
	b.fenceMark = ""
	return s
}

// String returns the current buffer contents (for tests/debug).
func (b *BufferManager) String() string {
	return b.buf.String()
}

// FormatProviderError is a lightweight helper used by callers to format provider errors.
func FormatProviderError(provider string, status int, body string) error {
	return fmt.Errorf("%s API returned status %d: %s", provider, status, body)
}
