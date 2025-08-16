package main

import (
	"fmt"
	"regexp"
	"strings"
)

// ANSI color codes for terminal formatting
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"
	
	// Colors
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	
	// Bright colors
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"
)

// MarkdownRenderer handles converting markdown to ANSI-formatted terminal output
type MarkdownRenderer struct {
	colorOutput bool
}

// NewMarkdownRenderer creates a new markdown renderer
func NewMarkdownRenderer(colorOutput bool) *MarkdownRenderer {
	return &MarkdownRenderer{
		colorOutput: colorOutput,
	}
}

// Render converts markdown text to ANSI-formatted terminal output
func (r *MarkdownRenderer) Render(markdown string) string {
	if !r.colorOutput {
		// If color is disabled, just clean up markdown syntax
		return r.stripMarkdown(markdown)
	}
	
	text := markdown
	
	// Process in order of specificity (most specific first)
	text = r.renderCodeBlocks(text)
	text = r.renderInlineCode(text)
	text = r.renderHeaders(text)
	text = r.renderBold(text)
	text = r.renderItalic(text)
	text = r.renderStrikethrough(text)
	text = r.renderLists(text)
	text = r.renderLinks(text)
	text = r.renderBlockquotes(text)
	
	return text
}

// renderHeaders converts markdown headers to ANSI formatted headers
func (r *MarkdownRenderer) renderHeaders(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	
	for _, line := range lines {
		// Handle ATX headers (# ## ###)
		if strings.HasPrefix(line, "#") {
			level := 0
			for i, char := range line {
				if char == '#' {
					level++
				} else if char == ' ' {
					break
				} else {
					level = 0
					break
				}
				if i >= 5 { // Max 6 levels
					break
				}
			}
			
			if level > 0 && level <= 6 {
				headerText := strings.TrimSpace(line[level:])
				switch level {
				case 1:
					result = append(result, fmt.Sprintf("%s%s%s%s", Bold, BrightRed, headerText, Reset))
				case 2:
					result = append(result, fmt.Sprintf("%s%s%s%s", Bold, BrightBlue, headerText, Reset))
				case 3:
					result = append(result, fmt.Sprintf("%s%s%s%s", Bold, BrightYellow, headerText, Reset))
				case 4:
					result = append(result, fmt.Sprintf("%s%s%s%s", Bold, BrightGreen, headerText, Reset))
				case 5:
					result = append(result, fmt.Sprintf("%s%s%s%s", Bold, BrightMagenta, headerText, Reset))
				case 6:
					result = append(result, fmt.Sprintf("%s%s%s%s", Bold, BrightCyan, headerText, Reset))
				}
				continue
			}
		}
		result = append(result, line)
	}
	
	return strings.Join(result, "\n")
}

// renderBold converts **text** and __text__ to ANSI bold
func (r *MarkdownRenderer) renderBold(text string) string {
	// Handle **text**
	re := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	text = re.ReplaceAllString(text, fmt.Sprintf("%s$1%s", Bold, Reset))
	
	// Handle __text__
	re = regexp.MustCompile(`__([^_]+)__`)
	text = re.ReplaceAllString(text, fmt.Sprintf("%s$1%s", Bold, Reset))
	
	return text
}

// renderItalic converts *text* and _text_ to ANSI italic
func (r *MarkdownRenderer) renderItalic(text string) string {
	// Handle *text* (but not **text**)
	re := regexp.MustCompile(`(?:^|[^*])\*([^*]+)\*(?:[^*]|$)`)
	text = re.ReplaceAllString(text, fmt.Sprintf("$1%s$2%s$3", Italic, Reset))
	
	// Handle _text_ (but not __text__)
	re = regexp.MustCompile(`(?:^|[^_])_([^_]+)_(?:[^_]|$)`)
	text = re.ReplaceAllString(text, fmt.Sprintf("$1%s$2%s$3", Italic, Reset))
	
	return text
}

// renderStrikethrough converts ~~text~~ to ANSI strikethrough (dim)
func (r *MarkdownRenderer) renderStrikethrough(text string) string {
	re := regexp.MustCompile(`~~([^~]+)~~`)
	text = re.ReplaceAllString(text, fmt.Sprintf("%s$1%s", Dim, Reset))
	return text
}

// renderInlineCode converts `code` to ANSI formatted code
func (r *MarkdownRenderer) renderInlineCode(text string) string {
	re := regexp.MustCompile("`([^`]+)`")
	text = re.ReplaceAllString(text, fmt.Sprintf("%s%s$1%s", Yellow, Bold, Reset))
	return text
}

// renderCodeBlocks converts ```code``` blocks to ANSI formatted code blocks
func (r *MarkdownRenderer) renderCodeBlocks(text string) string {
	// Handle fenced code blocks
	re := regexp.MustCompile("(?s)```([a-z]*)\n?(.*?)```")
	text = re.ReplaceAllStringFunc(text, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) >= 3 {
			language := parts[1]
			code := parts[2]
			
			// Add language label if present
			result := ""
			if language != "" {
				result += fmt.Sprintf("%s%s[%s]%s\n", Dim, Cyan, language, Reset)
			}
			
			// Format the code block
			codeLines := strings.Split(strings.TrimRight(code, "\n"), "\n")
			for _, line := range codeLines {
				result += fmt.Sprintf("%s%s%s%s\n", Dim, Yellow, line, Reset)
			}
			
			return result
		}
		return match
	})
	
	return text
}

// renderLists converts markdown lists to formatted lists
func (r *MarkdownRenderer) renderLists(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Unordered lists
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "+ ") {
			content := trimmed[2:]
			result = append(result, fmt.Sprintf("  %s•%s %s", BrightBlue, Reset, content))
			continue
		}
		
		// Ordered lists (basic detection)
		re := regexp.MustCompile(`^(\d+)\.\s+(.+)$`)
		if matches := re.FindStringSubmatch(trimmed); matches != nil {
			number := matches[1]
			content := matches[2]
			result = append(result, fmt.Sprintf("  %s%s.%s %s", BrightBlue, number, Reset, content))
			continue
		}
		
		result = append(result, line)
	}
	
	return strings.Join(result, "\n")
}

// renderLinks converts [text](url) to colored text
func (r *MarkdownRenderer) renderLinks(text string) string {
	re := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	text = re.ReplaceAllString(text, fmt.Sprintf("%s%s$1%s %s(%s$2%s)", Underline, BrightCyan, Reset, Dim, BrightCyan, Reset))
	return text
}

// renderBlockquotes converts > text to formatted blockquotes
func (r *MarkdownRenderer) renderBlockquotes(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "> ") {
			content := strings.TrimSpace(line)[2:]
			result = append(result, fmt.Sprintf("%s%s│%s %s", Dim, BrightBlue, Reset, content))
			continue
		}
		result = append(result, line)
	}
	
	return strings.Join(result, "\n")
}

// stripMarkdown removes markdown formatting when colors are disabled
func (r *MarkdownRenderer) stripMarkdown(text string) string {
	// Remove headers
	re := regexp.MustCompile(`^#{1,6}\s+`)
	text = re.ReplaceAllString(text, "")
	
	// Remove bold
	re = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	text = re.ReplaceAllString(text, "$1")
	re = regexp.MustCompile(`__([^_]+)__`)
	text = re.ReplaceAllString(text, "$1")
	
	// Remove italic
	re = regexp.MustCompile(`\*([^*]+)\*`)
	text = re.ReplaceAllString(text, "$1")
	re = regexp.MustCompile(`_([^_]+)_`)
	text = re.ReplaceAllString(text, "$1")
	
	// Remove strikethrough
	re = regexp.MustCompile(`~~([^~]+)~~`)
	text = re.ReplaceAllString(text, "$1")
	
	// Remove inline code
	re = regexp.MustCompile("`([^`]+)`")
	text = re.ReplaceAllString(text, "$1")
	
	// Remove code blocks
	re = regexp.MustCompile("(?s)```[a-z]*\n?(.*?)```")
	text = re.ReplaceAllString(text, "$1")
	
	// Remove links
	re = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	text = re.ReplaceAllString(text, "$1")
	
	// Clean up blockquotes
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "> ") {
			content := strings.TrimSpace(line)[2:]
			result = append(result, content)
		} else {
			result = append(result, line)
		}
	}
	text = strings.Join(result, "\n")
	
	return text
}

// StreamingMarkdownRenderer handles real-time markdown rendering for streaming output
type StreamingMarkdownRenderer struct {
	renderer *MarkdownRenderer
	buffer   strings.Builder
	inCodeBlock bool
	codeBlockLang string
}

// NewStreamingMarkdownRenderer creates a new streaming markdown renderer
func NewStreamingMarkdownRenderer(colorOutput bool) *StreamingMarkdownRenderer {
	return &StreamingMarkdownRenderer{
		renderer: NewMarkdownRenderer(colorOutput),
	}
}

// ProcessChunk processes a chunk of streaming text and returns formatted output
func (sr *StreamingMarkdownRenderer) ProcessChunk(chunk string) string {
	// Add chunk to buffer
	sr.buffer.WriteString(chunk)
	
	// Get the current buffer content
	bufferContent := sr.buffer.String()
	
	// For streaming, we need to be careful about partial markdown syntax
	// Only process complete lines or complete markdown constructs
	
	lines := strings.Split(bufferContent, "\n")
	
	// If we don't have a complete line (no newline at end), keep the last part in buffer
	if !strings.HasSuffix(bufferContent, "\n") && len(lines) > 1 {
		processableLines := lines[:len(lines)-1]
		remainingContent := lines[len(lines)-1]
		
		if len(processableLines) == 0 {
			return "" // Nothing to process yet
		}
		
		// Process the complete lines
		processableText := strings.Join(processableLines, "\n") + "\n"
		rendered := sr.renderer.Render(processableText)
		
		// Reset buffer to remaining content
		sr.buffer.Reset()
		if remainingContent != "" {
			sr.buffer.WriteString(remainingContent)
		}
		
		return rendered
	} else if strings.HasSuffix(bufferContent, "\n") {
		// We have complete content ending with newline
		rendered := sr.renderer.Render(bufferContent)
		sr.buffer.Reset()
		return rendered
	}
	
	// Buffer incomplete content
	return ""
}

// Flush processes any remaining content in the buffer
func (sr *StreamingMarkdownRenderer) Flush() string {
	if sr.buffer.Len() == 0 {
		return ""
	}
	
	remaining := sr.buffer.String()
	sr.buffer.Reset()
	
	return sr.renderer.Render(remaining)
}