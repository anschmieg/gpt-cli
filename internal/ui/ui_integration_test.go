//go:build integration
// +build integration

package ui

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestUI_Renderer_Integration(t *testing.T) {
    u := New()
    // Should always have a UI instance; renderer may be nil in some envs
    if u.Renderer != nil {
        out := u.Renderer.Render("# H\n\n**b**")
        assert.NotEmpty(t, out)
        // loosely assert header/body made it through
        assert.Contains(t, out, "H")
    }
}

