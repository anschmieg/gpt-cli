package ui

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

// UI handles the user interface styling and rendering
type UI struct {
	// Styles
	TitleStyle     lipgloss.Style
	SubtitleStyle  lipgloss.Style
	PromptStyle    lipgloss.Style
	InputStyle     lipgloss.Style
	ResponseStyle  lipgloss.Style
	ErrorStyle     lipgloss.Style
	LoadingStyle   lipgloss.Style
	HelpStyle      lipgloss.Style
	ContainerStyle lipgloss.Style

	// Fragment-aware renderer for streaming-safe rendering
	Renderer *Renderer
}

// New creates a new UI instance with default styles
func New() *UI {
	// Create fragment-aware renderer; enable TTY rendering only when stdout is a terminal
	var fragRenderer *Renderer
	if isatty.IsTerminal(os.Stdout.Fd()) {
		fragRenderer = NewRenderer(true)
	} else {
		fragRenderer = NewRenderer(false)
	}

	return &UI{
		TitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			MarginBottom(1),

		SubtitleStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			MarginBottom(1),

		PromptStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#059669")),

		InputStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#111827")).
			Background(lipgloss.Color("#F3F4F6")).
			Padding(0, 1).
			MarginTop(1),

		ResponseStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#111827")).
			Background(lipgloss.Color("#F9FAFB")).
			Padding(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#D1D5DB")).
			MarginTop(1).
			MarginBottom(1),

		ErrorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DC2626")).
			Background(lipgloss.Color("#FEF2F2")).
			Padding(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FCA5A5")).
			MarginTop(1).
			MarginBottom(1),

		LoadingStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true).
			MarginTop(2).
			MarginBottom(2),

		HelpStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			MarginTop(1),

		ContainerStyle: lipgloss.NewStyle().
			Padding(2),

		Renderer: fragRenderer,
	}
}

// Note: full-buffer Glamour rendering was replaced by the fragment-aware
// Renderer. The Renderer handles TTY vs non-TTY rendering and fragment-safe
// normalization; use it via `UI.Renderer`.

// IsMarkdown checks if text appears to contain markdown formatting
func (ui *UI) IsMarkdown(text string) bool {
	// Simple heuristics to detect markdown
	markdownIndicators := []string{
		"# ",   // Headers
		"## ",  // Headers
		"### ", // Headers
		"- ",   // Lists
		"* ",   // Lists
		"```",  // Code blocks
		"`",    // Inline code
		"**",   // Bold
		"__",   // Bold
		"*",    // Italic (but be careful of false positives)
		"_",    // Italic (but be careful of false positives)
		"[",    // Links
		"![",   // Images
	}

	for _, indicator := range markdownIndicators {
		if strings.Contains(text, indicator) {
			return true
		}
	}

	return false
}
