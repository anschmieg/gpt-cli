package ui

import (
	"strings"

	"github.com/charmbracelet/glamour"
)

// Renderer renders fragments for terminal consumption. When TTY is true this
// uses Glamour to convert markdown to ANSI; otherwise it returns plain text
// with newline normalization.
type Renderer struct {
	tty bool
	g   *glamour.TermRenderer
	// prevBlank indicates the last emitted fragment ended with a blank line.
	// This is used to avoid emitting duplicate blank lines when fragments are
	// rendered and printed separately by the CLI streaming loop.
	prevBlank bool
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
			return r.postProcess(s)
		}
		// fallthrough to plain
	}
	return r.postProcess(renderPlainNormalized(fragment))
}

// RenderPlain ensures newline-terminated plain fragments.
func RenderPlain(fragment string) string {
	return renderPlainNormalized(fragment)
}

// renderPlainNormalized trims per-line whitespace (except inside fenced code
// blocks), collapses multiple blank lines into a single blank line, and
// guarantees the output ends with exactly one newline.
func renderPlainNormalized(fragment string) string {
	if fragment == "" {
		return ""
	}
	lines := strings.Split(fragment, "\n")
	var out []string
	inFence := false
	for _, ln := range lines {
		trimmed := ln
		if strings.HasPrefix(strings.TrimSpace(ln), "```") {
			// fence delimiter toggles state; preserve the fence line as-is
			out = append(out, strings.TrimSpace(ln))
			inFence = !inFence
			continue
		}
		if inFence {
			// preserve indentation inside code fences
			out = append(out, ln)
			continue
		}
		// trim leading/trailing spaces for normal text
		trimmed = strings.TrimSpace(ln)
		out = append(out, trimmed)
	}

	// collapse multiple blank lines
	var collapsed []string
	blank := false
	for _, ln := range out {
		if ln == "" {
			if blank {
				// skip extra blank line
				continue
			}
			blank = true
			collapsed = append(collapsed, "")
			continue
		}
		blank = false
		collapsed = append(collapsed, ln)
	}

	// remove leading/trailing blank lines
	start := 0
	for start < len(collapsed) && collapsed[start] == "" {
		start++
	}
	end := len(collapsed)
	for end > start && collapsed[end-1] == "" {
		end--
	}

	result := strings.Join(collapsed[start:end], "\n")
	if result == "" {
		// represent an all-blank fragment as a single newline so downstream
		// postProcess can detect and collapse consecutive blank fragments.
		return "\n"
	}
	// ensure single trailing newline
	if !strings.HasSuffix(result, "\n") {
		result = result + "\n"
	}
	return result
}

// postProcess applies small stateful fixes across fragments:
//   - if the previous fragment ended with a blank line, drop leading blank
//     lines from this fragment to avoid consecutive empty lines when printing
//     fragments separately
//   - update r.prevBlank according to whether this fragment ends with a blank
//     line
func (r *Renderer) postProcess(s string) string {
	if s == "" {
		return ""
	}

	lines := strings.Split(s, "\n")
	var outLines []string
	inFence := false

	for _, ln := range lines {
		// detect fence start/end
		trimmed := strings.TrimSpace(ln)
		if strings.HasPrefix(trimmed, "```") {
			// toggle fence state and append the fence line as-is
			outLines = append(outLines, trimmed)
			inFence = !inFence
			continue
		}

		if inFence {
			// preserve code block lines exactly
			outLines = append(outLines, ln)
			continue
		}

		// non-code: collapse whitespace-only lines and trim others
		if strings.TrimSpace(ln) == "" {
			// if last appended line is blank, skip to avoid duplicates
			if len(outLines) > 0 && strings.TrimSpace(outLines[len(outLines)-1]) == "" {
				continue
			}
			outLines = append(outLines, "")
			continue
		}

		// trim leading/trailing spaces for normal text
		outLines = append(outLines, strings.TrimSpace(ln))
	}

	// Determine leading/trailing blank state
	firstBlank := len(outLines) > 0 && strings.TrimSpace(outLines[0]) == ""
	lastBlank := len(outLines) > 0 && strings.TrimSpace(outLines[len(outLines)-1]) == ""

	// compute start index: if previous fragment ended blank, drop leading blanks
	start := 0
	if r.prevBlank {
		for start < len(outLines) && strings.TrimSpace(outLines[start]) == "" {
			start++
		}
	} else {
		// keep at most one leading blank
		if firstBlank {
			start = 0
			// advance to first non-blank but leave a single blank at start
			i := 0
			for i < len(outLines) && strings.TrimSpace(outLines[i]) == "" {
				i++
			}
			if i > 1 {
				// collapse to single blank
				// shift slice so start points to the single blank
				// leave start as 0 and later collapse
			}
		}
	}

	// compute end index: trim trailing blanks
	end := len(outLines)
	for end > start && strings.TrimSpace(outLines[end-1]) == "" {
		end--
	}

	if start >= end {
		// nothing but blanks
		r.prevBlank = true
		return "\n"
	}

	segment := outLines[start:end]

	// compute minimal common indent of non-empty, non-list, non-blockquote lines
	minIndent := -1
	for _, ln := range segment {
		if strings.TrimSpace(ln) == "" {
			continue
		}
		t := strings.TrimLeft(ln, " \t")
		ts := strings.TrimSpace(t)
		// skip lists, blockquotes, fences and ANSI-looking lines
		if strings.HasPrefix(ts, "-") || strings.HasPrefix(ts, "*") || strings.HasPrefix(ts, ">") || strings.HasPrefix(ts, "1.") || strings.HasPrefix(t, "\x1b[") {
			continue
		}
		indent := len(ln) - len(t)
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}
	if minIndent > 0 {
		for i, ln := range segment {
			if len(ln) > minIndent {
				segment[i] = ln[minIndent:]
			}
		}
	}

	out := strings.Join(segment, "\n")

	// collapse runs of 3+ newlines defensively
	for strings.Contains(out, "\n\n\n") {
		out = strings.ReplaceAll(out, "\n\n\n", "\n\n")
	}

	// update prevBlank to whether original fragment ended with a blank line
	r.prevBlank = lastBlank

	return out + "\n"
}
