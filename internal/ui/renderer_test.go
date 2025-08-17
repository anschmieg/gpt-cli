package ui_test

import (
	"testing"

	"github.com/anschmieg/gpt-cli/internal/ui"
)

func TestRenderPlainAddsNewline(t *testing.T) {
	r := ui.NewRenderer(false)
	out := r.RenderFragment("hello")
	if out != "hello\n" {
		t.Fatalf("expected newline terminated output, got %q", out)
	}
}

func TestRendererNoPanicOnEmpty(t *testing.T) {
	r := ui.NewRenderer(false)
	if got := r.RenderFragment(""); got != "" {
		t.Fatalf("expected empty for empty input, got %q", got)
	}
}
