package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRendererNonTTYFragmentsMatchPlainJoin(t *testing.T) {
	// Use non-TTY renderer to get deterministic plain output
	r := NewRenderer(false)

	f1 := "prelude\n```\ncode line 1\n"
	f2 := "code line 2\n```\nafter\n"

	out1 := r.Render(f1)
	out2 := r.Render(f2)
	got := out1 + out2

	// Expected is the plain normalized rendering of the combined input
	want := RenderPlain(f1 + f2)

	assert.Equal(t, want, got)
}

func TestRendererTTYGracefulWhenGlamourUnavailable(t *testing.T) {
	// TTY renderer may or may not have Glamour available in the environment.
	// We assert both safe behaviors: if Glamour is unavailable, output must
	// equal the plain normalization; otherwise ensure important tokens exist.
	r := NewRenderer(true)

	f1 := "# Title\n\nSome text\n````go\nfunc main() {}\n````\n"
	f2 := "More text\n"

	out1 := r.Render(f1)
	out2 := r.Render(f2)
	got := out1 + out2

	if r.glamour == nil {
		// Glamour not initialized: deterministic plain output
		want := RenderPlain(f1 + f2)
		assert.Equal(t, want, got)
	} else {
		// Glamour initialized: output should include some rendered tokens
		assert.Contains(t, got, "Title")
		assert.Contains(t, got, "func main")
	}
}

func TestRendererFragmentBlankLineCollapse(t *testing.T) {
	// Ensure that when a fragment ends with a blank and the next begins
	// with blank lines, the renderer collapses them correctly.
	r := NewRenderer(false)

	f1 := "line one\n\n"   // ends with a blank line
	f2 := "\n\nline two\n" // starts with blank lines

	out1 := r.Render(f1)
	out2 := r.Render(f2)
	got := out1 + out2

	// The resulting output should contain a single blank line between
	// the two lines (renderer collapses duplicate blanks) and end with a newline.
	assert.Contains(t, got, "line one\nline two")
	assert.True(t, len(got) > 0 && got[len(got)-1] == '\n')
}
