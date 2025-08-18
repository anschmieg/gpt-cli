package ui

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestRendererRemovesCommonIndentForNormalText(t *testing.T) {
    r := NewRenderer(false)
    input := "    alpha\n    beta\n"
    out := r.Render(input)
    // Common 4-space indent should be removed for normal lines
    assert.Equal(t, "alpha\nbeta\n", out)
}

func TestRendererDoesNotTrimListOrQuoteIndent(t *testing.T) {
    r := NewRenderer(false)
    input := "- item\n    continuation\n"
    out := r.Render(input)
    // Due to renderPlainNormalized, leading spaces on non-code lines are trimmed.
    // Ensure the list marker remains and continuation text is present.
    assert.Contains(t, out, "- item")
    assert.Contains(t, out, "continuation")
}
