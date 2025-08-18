package modes

import (
	"fmt"
	"strings"
	"time"

	"github.com/anschmieg/gpt-cli/internal/config"
	"github.com/anschmieg/gpt-cli/internal/providers"
	"github.com/anschmieg/gpt-cli/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Message represents a single message in the conversation
type Message struct {
	Role      string    `json:"role"` // "user", "assistant", "system"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Conversation represents a chat conversation with memory
type Conversation struct {
	ID       string    `json:"id"`
	Messages []Message `json:"messages"`
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
}

// ChatMode handles chat functionality with conversation memory
type ChatMode struct {
	config       *config.Config
	provider     providers.Provider
	ui           *ui.UI
	conversation *Conversation
}

// ChatModel represents the BubbleTea model for chat mode
type ChatModel struct {
	chatMode      *ChatMode
	state         ChatState
	input         string
	cursor        int
	width         int
	height        int
	loading       bool
	error         string
	scrollOffset  int
	initialPrompt string
}

// ChatState represents the current state of the chat
type ChatState int

const (
	ChatStateInput ChatState = iota
	ChatStateLoading
	ChatStateResponse
	ChatStateError
)

// NewChatMode creates a new chat mode instance
func NewChatMode(config *config.Config, provider providers.Provider, ui *ui.UI) *ChatMode {
	conversation := &Conversation{
		ID:       generateConversationID(),
		Messages: []Message{},
		Created:  time.Now(),
		Updated:  time.Now(),
	}

	// Add system message if configured
	if config.System != "" {
		conversation.Messages = append(conversation.Messages, Message{
			Role:      "system",
			Content:   config.System,
			Timestamp: time.Now(),
		})
	}

	return &ChatMode{
		config:       config,
		provider:     provider,
		ui:           ui,
		conversation: conversation,
	}
}

// NewChatModel creates a new BubbleTea model for chat mode
func NewChatModel(chatMode *ChatMode, initial string) *ChatModel {
	return &ChatModel{
		chatMode:      chatMode,
		state:         ChatStateInput,
		initialPrompt: initial,
	}
}

// Init initializes the chat model
func (m *ChatModel) Init() tea.Cmd {
	cmds := []tea.Cmd{tea.EnterAltScreen}
	if strings.TrimSpace(m.initialPrompt) != "" {
		m.state = ChatStateLoading
		m.loading = true
		prompt := strings.TrimSpace(m.initialPrompt)
		m.initialPrompt = ""
		m.chatMode.conversation.Messages = append(m.chatMode.conversation.Messages, Message{
			Role:      "user",
			Content:   prompt,
			Timestamp: time.Now(),
		})
		m.chatMode.conversation.Updated = time.Now()
		cmds = append(cmds, m.makeRequest())
	}
	return tea.Batch(cmds...)
}

// Update handles messages and updates the chat model
func (m *ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case ResponseMsg:
		m.state = ChatStateResponse
		m.loading = false

		// Add assistant message to conversation
		m.chatMode.conversation.Messages = append(m.chatMode.conversation.Messages, Message{
			Role:      "assistant",
			Content:   msg.Content,
			Timestamp: time.Now(),
		})
		m.chatMode.conversation.Updated = time.Now()

		return m, nil

	case ErrorMsg:
		m.state = ChatStateError
		m.error = msg.Error
		m.loading = false
		return m, nil

	case StreamChunkMsg:
		// Handle streaming responses
		if len(m.chatMode.conversation.Messages) > 0 {
			lastMsg := &m.chatMode.conversation.Messages[len(m.chatMode.conversation.Messages)-1]
			if lastMsg.Role == "assistant" {
				lastMsg.Content += msg.Chunk
			} else {
				// Create new assistant message for streaming
				m.chatMode.conversation.Messages = append(m.chatMode.conversation.Messages, Message{
					Role:      "assistant",
					Content:   msg.Chunk,
					Timestamp: time.Now(),
				})
			}
		}
		return m, nil
	}

	return m, nil
}

// handleKeyMsg handles keyboard input in chat mode
func (m *ChatModel) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		if m.state == ChatStateInput || m.state == ChatStateResponse || m.state == ChatStateError {
			return m, tea.Quit
		}

	case "ctrl+l":
		// Clear conversation
		return m.clearConversation()

	case "ctrl+s":
		// Save conversation (future enhancement)
		return m, nil

	case "esc":
		if m.state == ChatStateResponse || m.state == ChatStateError {
			m.state = ChatStateInput
			m.error = ""
			return m, nil
		}

	case "enter":
		if m.state == ChatStateInput && strings.TrimSpace(m.input) != "" {
			return m.sendMessage()
		}

	case "backspace":
		if m.state == ChatStateInput && len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}

	case "up":
		if m.state == ChatStateResponse || m.state == ChatStateError {
			if m.scrollOffset > 0 {
				m.scrollOffset--
			}
		}

	case "down":
		if m.state == ChatStateResponse || m.state == ChatStateError {
			m.scrollOffset++
		}

	default:
		if m.state == ChatStateInput {
			m.input += msg.String()
		}
	}

	return m, nil
}

// sendMessage sends the current input as a user message
func (m *ChatModel) sendMessage() (tea.Model, tea.Cmd) {
	m.state = ChatStateLoading
	m.loading = true
	m.error = ""
	m.scrollOffset = 0

	prompt := strings.TrimSpace(m.input)
	m.input = ""

	// Add user message to conversation
	m.chatMode.conversation.Messages = append(m.chatMode.conversation.Messages, Message{
		Role:      "user",
		Content:   prompt,
		Timestamp: time.Now(),
	})
	m.chatMode.conversation.Updated = time.Now()

	return m, m.makeRequest()
}

// clearConversation clears the conversation history
func (m *ChatModel) clearConversation() (tea.Model, tea.Cmd) {
	m.chatMode.conversation.Messages = []Message{}

	// Re-add system message if configured
	if m.chatMode.config.System != "" {
		m.chatMode.conversation.Messages = append(m.chatMode.conversation.Messages, Message{
			Role:      "system",
			Content:   m.chatMode.config.System,
			Timestamp: time.Now(),
		})
	}

	m.chatMode.conversation.Updated = time.Now()
	m.state = ChatStateInput
	m.error = ""
	m.scrollOffset = 0

	return m, nil
}

// makeRequest creates a command to make a request to the provider
func (m *ChatModel) makeRequest() tea.Cmd {
	return func() tea.Msg {
		// Convert conversation messages to the format expected by the provider
		prompt := m.formatConversationForProvider()

		response, err := m.chatMode.provider.CallProvider(prompt)
		if err != nil {
			return ErrorMsg{Error: err.Error()}
		}
		return ResponseMsg{Content: response}
	}
}

// formatConversationForProvider formats the conversation for the provider
func (m *ChatModel) formatConversationForProvider() string {
	var parts []string

	for _, msg := range m.chatMode.conversation.Messages {
		switch msg.Role {
		case "system":
			if len(parts) == 0 {
				parts = append(parts, msg.Content)
			}
		case "user":
			parts = append(parts, fmt.Sprintf("User: %s", msg.Content))
		case "assistant":
			parts = append(parts, fmt.Sprintf("Assistant: %s", msg.Content))
		}
	}

	return strings.Join(parts, "\n\n")
}

// View renders the current chat view
func (m *ChatModel) View() string {
	switch m.state {
	case ChatStateInput:
		return m.renderChatView()
	case ChatStateLoading:
		return m.renderLoadingView()
	case ChatStateResponse:
		return m.renderChatView()
	case ChatStateError:
		return m.renderErrorView()
	default:
		return "Unknown state"
	}
}

// renderChatView renders the main chat interface
func (m *ChatModel) renderChatView() string {
	title := m.chatMode.ui.TitleStyle.Render("💬 Chat Mode")

	// Conversation info
	msgCount := len(m.chatMode.conversation.Messages)
	userMsgCount := 0
	for _, msg := range m.chatMode.conversation.Messages {
		if msg.Role == "user" {
			userMsgCount++
		}
	}

	info := m.chatMode.ui.SubtitleStyle.Render(fmt.Sprintf(
		"Messages: %d | User: %d | Provider: %s | Model: %s",
		msgCount,
		userMsgCount,
		m.chatMode.config.Provider,
		m.chatMode.config.Model,
	))

	// Chat history
	history := m.renderChatHistory()

	// Input area
	var inputArea string
	if m.loading {
		inputArea = m.chatMode.ui.LoadingStyle.Render("🤖 Thinking...")
	} else {
		prompt := m.chatMode.ui.PromptStyle.Render("You:")
		input := m.chatMode.ui.InputStyle.Render("> " + m.input + "█")
		inputArea = lipgloss.JoinVertical(lipgloss.Left, prompt, input)
	}

	// Help text
	help := m.chatMode.ui.HelpStyle.Render("Enter: Send • Ctrl+L: Clear • Ctrl+C/q: Quit • ↑↓: Scroll • Esc: Input mode")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		info,
		"",
		history,
		"",
		inputArea,
		"",
		help,
	)

	return m.chatMode.ui.ContainerStyle.Render(content)
}

// renderChatHistory renders the chat conversation history
func (m *ChatModel) renderChatHistory() string {
	if len(m.chatMode.conversation.Messages) == 0 {
		return m.chatMode.ui.SubtitleStyle.Render("No messages yet. Start the conversation!")
	}

	var messages []string

	// Apply scroll offset
	visibleMessages := m.chatMode.conversation.Messages
	if m.scrollOffset > 0 && m.scrollOffset < len(visibleMessages) {
		visibleMessages = visibleMessages[:len(visibleMessages)-m.scrollOffset]
	}

	for _, msg := range visibleMessages {
		if msg.Role == "system" {
			continue // Don't show system messages in chat history
		}

		var rendered string
		timestamp := msg.Timestamp.Format("15:04")

		switch msg.Role {
		case "user":
			header := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#0EA5E9")).
				Bold(true).
				Render(fmt.Sprintf("You (%s):", timestamp))
			content := msg.Content
			rendered = fmt.Sprintf("%s\n%s", header, content)

		case "assistant":
			header := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7C3AED")).
				Bold(true).
				Render(fmt.Sprintf("Assistant (%s):", timestamp))

			var content string
			if m.chatMode.config.Markdown && m.chatMode.ui.IsMarkdown(msg.Content) {
				content = m.chatMode.ui.RenderMarkdown(msg.Content)
			} else {
				content = msg.Content
			}
			rendered = fmt.Sprintf("%s\n%s", header, content)
		}

		if rendered != "" {
			messages = append(messages, rendered)
		}
	}

	if len(messages) == 0 {
		return m.chatMode.ui.SubtitleStyle.Render("No conversation history to display.")
	}

	historyContent := strings.Join(messages, "\n\n")

	// Wrap in a scrollable container
	return m.chatMode.ui.ResponseStyle.Render(historyContent)
}

// renderLoadingView renders the loading state
func (m *ChatModel) renderLoadingView() string {
	title := m.chatMode.ui.TitleStyle.Render("💬 Chat Mode")
	loading := m.chatMode.ui.LoadingStyle.Render("🤖 Assistant is thinking...")
	help := m.chatMode.ui.HelpStyle.Render("Press Ctrl+C to cancel")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		loading,
		"",
		help,
	)

	return m.chatMode.ui.ContainerStyle.Render(content)
}

// renderErrorView renders the error state
func (m *ChatModel) renderErrorView() string {
	title := m.chatMode.ui.TitleStyle.Render("💬 Chat Mode - Error")
	error := m.chatMode.ui.ErrorStyle.Render("❌ " + m.error)
	help := m.chatMode.ui.HelpStyle.Render("Press Esc to continue chatting • Ctrl+C/q to quit")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		error,
		"",
		help,
	)

	return m.chatMode.ui.ContainerStyle.Render(content)
}

// generateConversationID generates a unique ID for the conversation
func generateConversationID() string {
	// Use nanosecond precision to avoid collisions in quick succession.
	// This eliminates the need for explicit delays in tests or calling code.
	return fmt.Sprintf("chat_%d", time.Now().UnixNano())
}

// Message types for BubbleTea
type ResponseMsg struct {
	Content string
}

type ErrorMsg struct {
	Error string
}

type StreamChunkMsg struct {
	Chunk string
}

// GetConversation returns the current conversation
func (c *ChatMode) GetConversation() *Conversation {
	return c.conversation
}

// LoadConversation loads a conversation from history
func (c *ChatMode) LoadConversation(conversation *Conversation) {
	c.conversation = conversation
}

// ExportConversation exports the conversation in a readable format
func (c *ChatMode) ExportConversation() string {
	var parts []string

	parts = append(parts, fmt.Sprintf("# Conversation %s", c.conversation.ID))
	parts = append(parts, fmt.Sprintf("Created: %s", c.conversation.Created.Format("2006-01-02 15:04:05")))
	parts = append(parts, fmt.Sprintf("Updated: %s", c.conversation.Updated.Format("2006-01-02 15:04:05")))
	parts = append(parts, "")

	for _, msg := range c.conversation.Messages {
		if msg.Role == "system" {
			continue
		}

		var role string
		switch msg.Role {
		case "user":
			role = "**You**"
		case "assistant":
			role = "**Assistant**"
		}

		timestamp := msg.Timestamp.Format("15:04")
		parts = append(parts, fmt.Sprintf("%s (%s):\n%s\n", role, timestamp, msg.Content))
	}

	return strings.Join(parts, "\n")
}
