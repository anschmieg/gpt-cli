package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	ui := New()

	assert.NotNil(t, ui)
	assert.NotNil(t, ui.TitleStyle)
	assert.NotNil(t, ui.SubtitleStyle)
	assert.NotNil(t, ui.PromptStyle)
	assert.NotNil(t, ui.InputStyle)
	assert.NotNil(t, ui.ResponseStyle)
	assert.NotNil(t, ui.ErrorStyle)
	assert.NotNil(t, ui.LoadingStyle)
	assert.NotNil(t, ui.HelpStyle)
	assert.NotNil(t, ui.ContainerStyle)
	assert.NotNil(t, ui.glamourRenderer)
}

func TestRenderMarkdown(t *testing.T) {
	ui := New()

	tests := []struct {
		name     string
		input    string
		contains []string // strings that should be present in rendered output
	}{
		{
			name:     "plain text",
			input:    "Hello world",
			contains: []string{"Hello world"},
		},
		{
			name:     "header",
			input:    "# Main Title\n\nContent here",
			contains: []string{"Main Title", "Content here"},
		},
		{
			name:     "code block",
			input:    "```go\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```",
			contains: []string{"func main()", "fmt.Println"},
		},
		{
			name:     "list",
			input:    "- Item 1\n- Item 2\n- Item 3",
			contains: []string{"Item 1", "Item 2", "Item 3"},
		},
		{
			name:     "inline code",
			input:    "Use `fmt.Println` to print",
			contains: []string{"fmt.Println", "print"},
		},
		{
			name:     "bold text",
			input:    "This is **bold** text",
			contains: []string{"bold", "text"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ui.RenderMarkdown(tt.input)

			// Check that result is not empty
			assert.NotEmpty(t, result)

			// Check that expected strings are present
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected, "Expected '%s' to be in rendered output", expected)
			}
		})
	}
}

func TestIsMarkdown(t *testing.T) {
	ui := New()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "plain text",
			input:    "Hello world",
			expected: false,
		},
		{
			name:     "header level 1",
			input:    "# Title",
			expected: true,
		},
		{
			name:     "header level 2",
			input:    "## Subtitle",
			expected: true,
		},
		{
			name:     "header level 3",
			input:    "### Section",
			expected: true,
		},
		{
			name:     "unordered list with dash",
			input:    "- Item 1",
			expected: true,
		},
		{
			name:     "unordered list with asterisk",
			input:    "* Item 1",
			expected: true,
		},
		{
			name:     "code block",
			input:    "```go\ncode here\n```",
			expected: true,
		},
		{
			name:     "inline code",
			input:    "Use `code` here",
			expected: true,
		},
		{
			name:     "bold with asterisks",
			input:    "This is **bold**",
			expected: true,
		},
		{
			name:     "bold with underscores",
			input:    "This is __bold__",
			expected: true,
		},
		{
			name:     "italic with asterisk",
			input:    "This is *italic*",
			expected: true,
		},
		{
			name:     "italic with underscore",
			input:    "This is _italic_",
			expected: true,
		},
		{
			name:     "link",
			input:    "Check [this link](http://example.com)",
			expected: true,
		},
		{
			name:     "image",
			input:    "![Alt text](image.png)",
			expected: true,
		},
		{
			name:     "false positive asterisk",
			input:    "2 * 3 = 6",
			expected: true, // This will be detected as markdown, which is acceptable
		},
		{
			name:     "false positive underscore",
			input:    "file_name.txt",
			expected: true, // This will be detected as markdown, which is acceptable
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ui.IsMarkdown(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRenderMarkdownFallback(t *testing.T) {
	// Test fallback behavior when renderer is nil
	ui := &UI{glamourRenderer: nil}

	input := "# Test Header\nSome content"
	result := ui.RenderMarkdown(input)

	// Should return original text when renderer is nil
	assert.Equal(t, input, result)
}
