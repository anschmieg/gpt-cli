package ui

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
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
}

// New creates a new UI instance with default styles
func New() *UI {
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
	}
}

// RenderMarkdown renders markdown text with basic formatting
func (ui *UI) RenderMarkdown(text string) string {
	if !hasMarkdown(text) {
		return text
	}

	lines := strings.Split(text, "\n")
	var result []string

	inCodeBlock := false
	for _, line := range lines {
		// Handle code blocks
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCodeBlock = !inCodeBlock
			if inCodeBlock {
				result = append(result, ui.codeBlockStyle("--- code ---"))
			} else {
				result = append(result, ui.codeBlockStyle("--- end code ---"))
			}
			continue
		}

		if inCodeBlock {
			result = append(result, ui.codeBlockStyle(line))
			continue
		}

		// Handle headers
		if match := regexp.MustCompile(`^(#{1,6})\s+(.*)$`).FindStringSubmatch(line); match != nil {
			level := len(match[1])
			text := match[2]
			result = append(result, ui.headerStyle(text, level))
			continue
		}

		// Handle lists
		if match := regexp.MustCompile(`^\s*([-*+])\s+(.*)$`).FindStringSubmatch(line); match != nil {
			text := match[2]
			result = append(result, ui.listStyle(text))
			continue
		}

		// Handle inline formatting
		processed := line
		processed = ui.processInlineCode(processed)
		processed = ui.processBold(processed)
		processed = ui.processItalic(processed)

		result = append(result, processed)
	}

	return strings.Join(result, "\n")
}

// hasMarkdown checks if text contains markdown formatting
func hasMarkdown(text string) bool {
	patterns := []string{
		`^#{1,6}\s+`,     // Headers
		`^\s*[-*+]\s+`,   // Lists
		"```",            // Code blocks
		"`[^`]+`",        // Inline code
		`\*\*[^*]+\*\*`,  // Bold
		`\*[^*]+\*`,      // Italic
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			return true
		}
	}
	return false
}

// Style helpers
func (ui *UI) headerStyle(text string, level int) string {
	style := lipgloss.NewStyle().Bold(true)
	
	if level <= 2 {
		style = style.Foreground(lipgloss.Color("#0EA5E9")) // Cyan
	} else {
		style = style.Foreground(lipgloss.Color("#EAB308")) // Yellow
	}
	
	return style.Render(text)
}

func (ui *UI) listStyle(text string) string {
	return "  • " + text
}

func (ui *UI) codeBlockStyle(text string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Render(text)
}

func (ui *UI) processInlineCode(text string) string {
	re := regexp.MustCompile("`([^`]+)`")
	return re.ReplaceAllStringFunc(text, func(match string) string {
		code := strings.Trim(match, "`")
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#059669")).
			Render(code)
	})
}

func (ui *UI) processBold(text string) string {
	re := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		content := strings.Trim(match, "*")
		return lipgloss.NewStyle().
			Bold(true).
			Render(content)
	})
}

func (ui *UI) processItalic(text string) string {
	re := regexp.MustCompile(`\*([^*]+)\*`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		content := strings.Trim(match, "*")
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EAB308")).
			Render(content)
	})
}