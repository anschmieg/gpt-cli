package ui

import (
	"github.com/charmbracelet/glamour"
)

// Renderer renders fragments for terminal consumption. When TTY is true this
// uses Glamour to convert markdown to ANSI; otherwise it returns plain text
// with newline normalization.
type Renderer struct {
	tty bool
	g   *glamour.TermRenderer
}

// NewRenderer constructs a Renderer and attempts to create a Glamour term
// renderer when tty is true. If Glamour initialization fails we gracefully
// fall back to plain text rendering.
func NewRenderer(tty bool) *Renderer {
	r := &Renderer{tty: tty}
	if tty {
		if tr, err := glamour.NewTermRenderer(glamour.WithAutoStyle()); err == nil {
			r.g = tr
		}
	}
	return r
}

// RenderFragment renders a markdown fragment. If a Glamour renderer is
// available it will be used; otherwise the fragment is returned with a
// trailing newline.
func (r *Renderer) RenderFragment(fragment string) string {
	if r == nil || fragment == "" {
		return ""
	}
	if r.g != nil {
		if s, err := r.g.Render(fragment); err == nil {
			return s
		}
		// fallthrough to plain
	}
	if fragment[len(fragment)-1] != '\n' {
		return fragment + "\n"
	}
	return fragment
}

// RenderPlain ensures newline-terminated plain fragments.
func RenderPlain(fragment string) string {
	if fragment == "" {
		return ""
	}
	if fragment[len(fragment)-1] != '\n' {
		return fragment + "\n"
	}
	return fragment
}
