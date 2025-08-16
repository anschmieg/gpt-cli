package main

import (
	"strings"
	"testing"
)

func TestMarkdownRenderer_RenderHeaders(t *testing.T) {
	renderer := NewMarkdownRenderer(true)
	
	tests := []struct {
		name     string
		input    string
		contains []string
		notContains []string
	}{
		{
			name:     "h1_header",
			input:    "# Main Title",
			contains: []string{"Main Title", Bold, BrightRed},
			notContains: []string{"#"},
		},
		{
			name:     "h2_header",
			input:    "## Subtitle",
			contains: []string{"Subtitle", Bold, BrightBlue},
			notContains: []string{"##"},
		},
		{
			name:     "h3_header",
			input:    "### Section",
			contains: []string{"Section", Bold, BrightYellow},
			notContains: []string{"###"},
		},
		{
			name:     "not_a_header",
			input:    "This # is not a header",
			contains: []string{"This # is not a header"},
			notContains: []string{Bold},
		},
		{
			name:     "multiple_headers",
			input:    "# Title\n## Subtitle\nRegular text",
			contains: []string{"Title", "Subtitle", "Regular text", Bold},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.Render(tt.input)
			
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain '%s', got: %s", expected, result)
				}
			}
			
			for _, notExpected := range tt.notContains {
				if strings.Contains(result, notExpected) {
					t.Errorf("Expected result NOT to contain '%s', got: %s", notExpected, result)
				}
			}
		})
	}
}

func TestMarkdownRenderer_RenderBold(t *testing.T) {
	renderer := NewMarkdownRenderer(true)
	
	tests := []struct {
		name   string
		input  string
		expected string
	}{
		{
			name:   "double_asterisk",
			input:  "This is **bold** text",
			expected: "This is " + Bold + "bold" + Reset + " text",
		},
		{
			name:   "double_underscore",
			input:  "This is __bold__ text",
			expected: "This is " + Bold + "bold" + Reset + " text",
		},
		{
			name:   "multiple_bold",
			input:  "**First** and **second** bold",
			expected: Bold + "First" + Reset + " and " + Bold + "second" + Reset + " bold",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.renderBold(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestMarkdownRenderer_RenderInlineCode(t *testing.T) {
	renderer := NewMarkdownRenderer(true)
	
	input := "Use `ls -la` to list files"
	result := renderer.renderInlineCode(input)
	
	expected := "Use " + Yellow + Bold + "ls -la" + Reset + " to list files"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestMarkdownRenderer_RenderCodeBlocks(t *testing.T) {
	renderer := NewMarkdownRenderer(true)
	
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name: "simple_code_block",
			input: "```\necho hello\nls -la\n```",
			contains: []string{"echo hello", "ls -la", Yellow, Dim},
		},
		{
			name: "code_block_with_language",
			input: "```bash\necho hello\n```",
			contains: []string{"echo hello", "[bash]", Yellow, Cyan},
		},
		{
			name: "multiple_code_blocks",
			input: "```\nfirst\n```\n\nText\n\n```\nsecond\n```",
			contains: []string{"first", "second", "Text"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.renderCodeBlocks(tt.input)
			
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain '%s', got: %s", expected, result)
				}
			}
		})
	}
}

func TestMarkdownRenderer_RenderLists(t *testing.T) {
	renderer := NewMarkdownRenderer(true)
	
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name: "unordered_list_dash",
			input: "- First item\n- Second item",
			contains: []string{"•", "First item", "Second item", BrightBlue},
		},
		{
			name: "unordered_list_asterisk",
			input: "* First item\n* Second item",
			contains: []string{"•", "First item", "Second item"},
		},
		{
			name: "ordered_list",
			input: "1. First item\n2. Second item",
			contains: []string{"1.", "2.", "First item", "Second item"},
		},
		{
			name: "mixed_content",
			input: "Normal text\n- List item\nMore text",
			contains: []string{"Normal text", "•", "List item", "More text"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.renderLists(tt.input)
			
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain '%s', got: %s", expected, result)
				}
			}
		})
	}
}

func TestMarkdownRenderer_RenderLinks(t *testing.T) {
	renderer := NewMarkdownRenderer(true)
	
	input := "Visit [GitHub](https://github.com) for more info"
	result := renderer.renderLinks(input)
	
	// Should contain the link text with formatting and the URL
	if !strings.Contains(result, "GitHub") {
		t.Error("Should contain link text 'GitHub'")
	}
	
	if !strings.Contains(result, "https://github.com") {
		t.Error("Should contain URL")
	}
	
	if !strings.Contains(result, Underline) {
		t.Error("Should contain underline formatting")
	}
	
	if !strings.Contains(result, BrightCyan) {
		t.Error("Should contain cyan color")
	}
}

func TestMarkdownRenderer_RenderBlockquotes(t *testing.T) {
	renderer := NewMarkdownRenderer(true)
	
	input := "> This is a quote\n> Second line\nNormal text"
	result := renderer.renderBlockquotes(input)
	
	// Should contain blockquote formatting
	if !strings.Contains(result, "│") {
		t.Error("Should contain blockquote indicator")
	}
	
	if !strings.Contains(result, "This is a quote") {
		t.Error("Should contain quote text")
	}
	
	if !strings.Contains(result, "Normal text") {
		t.Error("Should contain normal text")
	}
}

func TestMarkdownRenderer_StripMarkdown(t *testing.T) {
	renderer := NewMarkdownRenderer(false) // No colors
	
	input := "# Header\n**Bold** and *italic* and `code` and [link](url)\n> Quote"
	result := renderer.stripMarkdown(input)
	
	// Should not contain markdown syntax
	notExpected := []string{"#", "**", "*", "`", "[", "]", "(", ")", ">"}
	for _, syntax := range notExpected {
		if strings.Contains(result, syntax) {
			t.Errorf("Stripped markdown should not contain '%s', got: %s", syntax, result)
		}
	}
	
	// Should contain the actual text
	expected := []string{"Header", "Bold", "italic", "code", "link", "Quote"}
	for _, text := range expected {
		if !strings.Contains(result, text) {
			t.Errorf("Stripped markdown should contain '%s', got: %s", text, result)
		}
	}
}

func TestStreamingMarkdownRenderer(t *testing.T) {
	// Test processing chunks
	tests := []struct {
		name   string
		chunks []string
		expectOutput bool
	}{
		{
			name:   "partial_line",
			chunks: []string{"Hello "},
			expectOutput: false, // Should buffer partial line
		},
		{
			name:   "complete_line",
			chunks: []string{"Hello world\n"},
			expectOutput: true,
		},
		{
			name:   "multiple_chunks",
			chunks: []string{"Hello ", "world\n", "Next line\n"},
			expectOutput: true,
		},
		{
			name:   "markdown_formatting",
			chunks: []string{"**Bold** text\n"},
			expectOutput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			streamRenderer := NewStreamingMarkdownRenderer(true)
			var hasOutput bool
			
			for _, chunk := range tt.chunks {
				result := streamRenderer.ProcessChunk(chunk)
				if result != "" {
					hasOutput = true
				}
			}
			
			if hasOutput != tt.expectOutput {
				t.Errorf("Expected output: %v, got output: %v", tt.expectOutput, hasOutput)
			}
		})
	}
}

func TestStreamingMarkdownRenderer_Flush(t *testing.T) {
	renderer := NewStreamingMarkdownRenderer(true)
	
	// Add some content without newline
	renderer.ProcessChunk("Hello world")
	
	// Should be buffered, no output yet
	result := renderer.ProcessChunk("")
	if result != "" {
		t.Error("Should not output incomplete line")
	}
	
	// Flush should return the buffered content
	flushed := renderer.Flush()
	if flushed == "" {
		t.Error("Flush should return buffered content")
	}
	
	if !strings.Contains(flushed, "Hello world") {
		t.Error("Flushed content should contain the text")
	}
	
	// Second flush should return empty
	flushed2 := renderer.Flush()
	if flushed2 != "" {
		t.Error("Second flush should return empty string")
	}
}

func TestMarkdownRendererColorDisabled(t *testing.T) {
	renderer := NewMarkdownRenderer(false) // Disabled colors
	
	input := "# Header\n**Bold** text with `code`"
	result := renderer.Render(input)
	
	// Should not contain ANSI codes
	if strings.Contains(result, "\033[") {
		t.Error("Result should not contain ANSI color codes when colors disabled")
	}
	
	// Should still contain the text content
	if !strings.Contains(result, "Header") {
		t.Error("Should contain header text")
	}
	
	if !strings.Contains(result, "Bold") {
		t.Error("Should contain bold text")
	}
	
	if !strings.Contains(result, "code") {
		t.Error("Should contain code text")
	}
}

func TestMarkdownEdgeCases(t *testing.T) {
	renderer := NewMarkdownRenderer(true)
	
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty_input",
			input: "",
		},
		{
			name:  "only_newlines",
			input: "\n\n\n",
		},
		{
			name:  "unclosed_markdown",
			input: "**bold without closing",
		},
		{
			name:  "nested_markdown",
			input: "**bold with *italic* inside**",
		},
		{
			name:  "multiple_code_blocks",
			input: "```\nfirst\n```\n\n```bash\nsecond\n```",
		},
		{
			name:  "complex_lists",
			input: "- Item 1\n  - Nested item\n- Item 2\n1. Ordered\n2. List",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			result := renderer.Render(tt.input)
			_ = result // Use the result to avoid unused variable
		})
	}
}