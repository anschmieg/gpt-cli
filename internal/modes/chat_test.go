package modes

import (
	"testing"
	"time"

	"github.com/anschmieg/gpt-cli/internal/config"
	"github.com/anschmieg/gpt-cli/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewChatMode(t *testing.T) {
	cfg := &config.Config{
		Provider: "mock",
		Model:    "test-model",
		System:   "You are a helpful assistant",
	}
	provider := &MockProvider{}
	ui := ui.New()

	mode := NewChatMode(cfg, provider, ui)

	assert.NotNil(t, mode)
	assert.Equal(t, cfg, mode.config)
	assert.Equal(t, provider, mode.provider)
	assert.Equal(t, ui, mode.ui)
	assert.NotNil(t, mode.conversation)

	// Should have system message
	require.Len(t, mode.conversation.Messages, 1)
	assert.Equal(t, "system", mode.conversation.Messages[0].Role)
	assert.Equal(t, "You are a helpful assistant", mode.conversation.Messages[0].Content)
}

func TestNewChatModel(t *testing.T) {
	cfg := &config.Config{Provider: "mock", Model: "test-model"}
	provider := &MockProvider{}
	ui := ui.New()

	chatMode := NewChatMode(cfg, provider, ui)
	model := NewChatModel(chatMode, "")

	assert.NotNil(t, model)
	assert.Equal(t, chatMode, model.chatMode)
	assert.Equal(t, ChatStateInput, model.state)
	assert.Empty(t, model.input)
	assert.False(t, model.loading)
	assert.Empty(t, model.error)
}

func TestChatModelInit(t *testing.T) {
	cfg := &config.Config{Provider: "mock", Model: "test-model"}
	provider := &MockProvider{}
	ui := ui.New()

	chatMode := NewChatMode(cfg, provider, ui)
	model := NewChatModel(chatMode, "")

	cmd := model.Init()
	assert.NotNil(t, cmd) // Just check that a command is returned
}

func TestChatModelInitWithInitialPrompt(t *testing.T) {
	cfg := &config.Config{Provider: "mock", Model: "test-model"}
	provider := &MockProvider{response: "Hello!"}
	ui := ui.New()

	chatMode := NewChatMode(cfg, provider, ui)
	model := NewChatModel(chatMode, "Hi")

	cmd := model.Init()
	require.NotNil(t, cmd)

	// Should have user message added
	require.Len(t, chatMode.conversation.Messages, 1)
	assert.Equal(t, "user", chatMode.conversation.Messages[0].Role)
	assert.Equal(t, ChatStateLoading, model.state)
	assert.True(t, model.loading)
}

func TestChatModelUpdate(t *testing.T) {
	cfg := &config.Config{Provider: "mock", Model: "test-model"}
	provider := &MockProvider{}
	ui := ui.New()

	chatMode := NewChatMode(cfg, provider, ui)
	model := NewChatModel(chatMode, "")

	t.Run("window size message", func(t *testing.T) {
		msg := tea.WindowSizeMsg{Width: 80, Height: 24}
		updatedModel, cmd := model.Update(msg)

		assert.Nil(t, cmd)
		chatModel := updatedModel.(*ChatModel)
		assert.Equal(t, 80, chatModel.width)
		assert.Equal(t, 24, chatModel.height)
	})

	t.Run("response message", func(t *testing.T) {
		// First add a user message
		model.chatMode.conversation.Messages = append(model.chatMode.conversation.Messages, Message{
			Role:      "user",
			Content:   "Hello",
			Timestamp: time.Now(),
		})

		msg := ResponseMsg{Content: "Hello! How can I help you?"}
		updatedModel, cmd := model.Update(msg)

		assert.Nil(t, cmd)
		chatModel := updatedModel.(*ChatModel)
		assert.Equal(t, ChatStateResponse, chatModel.state)
		assert.False(t, chatModel.loading)

		// Should have added assistant message
		messages := chatModel.chatMode.conversation.Messages
		require.Len(t, messages, 2) // user + assistant
		assert.Equal(t, "assistant", messages[1].Role)
		assert.Equal(t, "Hello! How can I help you?", messages[1].Content)
	})

	t.Run("error message", func(t *testing.T) {
		msg := ErrorMsg{Error: "Something went wrong"}
		updatedModel, cmd := model.Update(msg)

		assert.Nil(t, cmd)
		chatModel := updatedModel.(*ChatModel)
		assert.Equal(t, ChatStateError, chatModel.state)
		assert.Equal(t, "Something went wrong", chatModel.error)
		assert.False(t, chatModel.loading)
	})

	t.Run("stream chunk message", func(t *testing.T) {
		// Add assistant message first
		model.chatMode.conversation.Messages = append(model.chatMode.conversation.Messages, Message{
			Role:      "assistant",
			Content:   "Hello",
			Timestamp: time.Now(),
		})

		msg := StreamChunkMsg{Chunk: " world!"}
		updatedModel, cmd := model.Update(msg)

		assert.Nil(t, cmd)
		chatModel := updatedModel.(*ChatModel)

		// Should have appended to last assistant message
		messages := chatModel.chatMode.conversation.Messages
		lastMsg := messages[len(messages)-1]
		assert.Equal(t, "assistant", lastMsg.Role)
		assert.Equal(t, "Hello world!", lastMsg.Content)
	})
}

func TestChatModelKeyHandling(t *testing.T) {
	cfg := &config.Config{Provider: "mock", Model: "test-model"}
	provider := &MockProvider{response: "Test response"}
	ui := ui.New()

	chatMode := NewChatMode(cfg, provider, ui)
	model := NewChatModel(chatMode, "")

	t.Run("typing input", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}
		updatedModel, cmd := model.Update(msg)

		assert.Nil(t, cmd)
		chatModel := updatedModel.(*ChatModel)
		assert.Equal(t, "h", chatModel.input)
	})

	t.Run("backspace", func(t *testing.T) {
		model.input = "hello"
		msg := tea.KeyMsg{Type: tea.KeyBackspace}
		updatedModel, cmd := model.Update(msg)

		assert.Nil(t, cmd)
		chatModel := updatedModel.(*ChatModel)
		assert.Equal(t, "hell", chatModel.input)
	})

	t.Run("enter with input", func(t *testing.T) {
		model.input = "test message"
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, cmd := model.Update(msg)

		assert.NotNil(t, cmd)
		chatModel := updatedModel.(*ChatModel)
		assert.Equal(t, ChatStateLoading, chatModel.state)
		assert.True(t, chatModel.loading)
		assert.Empty(t, chatModel.input) // Should be cleared

		// Should have added user message
		messages := chatModel.chatMode.conversation.Messages
		lastMsg := messages[len(messages)-1]
		assert.Equal(t, "user", lastMsg.Role)
		assert.Equal(t, "test message", lastMsg.Content)
	})

	t.Run("clear conversation", func(t *testing.T) {
		// Add some messages
		model.chatMode.conversation.Messages = append(model.chatMode.conversation.Messages,
			Message{Role: "user", Content: "Hello", Timestamp: time.Now()},
			Message{Role: "assistant", Content: "Hi there!", Timestamp: time.Now()},
		)

		msg := tea.KeyMsg{Type: tea.KeyCtrlL}
		updatedModel, cmd := model.Update(msg)

		assert.Nil(t, cmd)
		chatModel := updatedModel.(*ChatModel)
		assert.Equal(t, ChatStateInput, chatModel.state)
		assert.Empty(t, chatModel.error)

		// Should have cleared messages (except system message if present)
		assert.Len(t, chatModel.chatMode.conversation.Messages, 0)
	})

	t.Run("quit", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyCtrlC}
		updatedModel, cmd := model.Update(msg)

		assert.NotNil(t, cmd) // Should return a quit command
		assert.NotNil(t, updatedModel)
	})

	t.Run("escape from error state", func(t *testing.T) {
		model.state = ChatStateError
		model.error = "Some error"

		msg := tea.KeyMsg{Type: tea.KeyEsc}
		updatedModel, cmd := model.Update(msg)

		assert.Nil(t, cmd)
		chatModel := updatedModel.(*ChatModel)
		assert.Equal(t, ChatStateInput, chatModel.state)
		assert.Empty(t, chatModel.error)
	})
}

func TestFormatConversationForProvider(t *testing.T) {
	cfg := &config.Config{Provider: "mock", Model: "test-model", System: "You are helpful"}
	provider := &MockProvider{}
	ui := ui.New()

	chatMode := NewChatMode(cfg, provider, ui)
	model := NewChatModel(chatMode, "")

	// Add some messages
	model.chatMode.conversation.Messages = append(model.chatMode.conversation.Messages,
		Message{Role: "user", Content: "Hello", Timestamp: time.Now()},
		Message{Role: "assistant", Content: "Hi there!", Timestamp: time.Now()},
		Message{Role: "user", Content: "How are you?", Timestamp: time.Now()},
	)

	formatted := model.formatConversationForProvider()

	expected := "You are helpful\n\nUser: Hello\n\nAssistant: Hi there!\n\nUser: How are you?"
	assert.Equal(t, expected, formatted)
}

func TestGenerateConversationID(t *testing.T) {
	id1 := generateConversationID()
	time.Sleep(1 * time.Second) // Ensure different timestamp
	id2 := generateConversationID()

	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "chat_")
	assert.Contains(t, id2, "chat_")

	// Test that they both follow expected format
	assert.Regexp(t, `^chat_\d+$`, id1)
	assert.Regexp(t, `^chat_\d+$`, id2)
}

func TestConversationMethods(t *testing.T) {
	cfg := &config.Config{Provider: "mock", Model: "test-model"}
	provider := &MockProvider{}
	ui := ui.New()

	chatMode := NewChatMode(cfg, provider, ui)

	t.Run("GetConversation", func(t *testing.T) {
		conv := chatMode.GetConversation()
		assert.NotNil(t, conv)
		assert.Equal(t, chatMode.conversation, conv)
	})

	t.Run("LoadConversation", func(t *testing.T) {
		newConv := &Conversation{
			ID:       "test_conv",
			Messages: []Message{{Role: "user", Content: "Test", Timestamp: time.Now()}},
			Created:  time.Now(),
			Updated:  time.Now(),
		}

		chatMode.LoadConversation(newConv)
		assert.Equal(t, newConv, chatMode.conversation)
	})

	t.Run("ExportConversation", func(t *testing.T) {
		// Add some messages
		chatMode.conversation.Messages = []Message{
			{Role: "user", Content: "Hello", Timestamp: time.Now()},
			{Role: "assistant", Content: "Hi there!", Timestamp: time.Now()},
		}

		exported := chatMode.ExportConversation()

		assert.Contains(t, exported, "# Conversation")
		assert.Contains(t, exported, "Created:")
		assert.Contains(t, exported, "Updated:")
		assert.Contains(t, exported, "**You**")
		assert.Contains(t, exported, "**Assistant**")
		assert.Contains(t, exported, "Hello")
		assert.Contains(t, exported, "Hi there!")
	})
}

func TestChatModelView(t *testing.T) {
	cfg := &config.Config{Provider: "mock", Model: "test-model"}
	provider := &MockProvider{}
	ui := ui.New()

	chatMode := NewChatMode(cfg, provider, ui)
	model := NewChatModel(chatMode, "")

	// Set some dimensions
	model.width = 80
	model.height = 24

	t.Run("input state", func(t *testing.T) {
		model.state = ChatStateInput
		view := model.View()

		assert.Contains(t, view, "💬 Chat Mode")
		assert.Contains(t, view, "Enter: Send")
		assert.Contains(t, view, "Ctrl+C/q: Quit")
	})

	t.Run("loading state", func(t *testing.T) {
		model.state = ChatStateLoading
		model.loading = true
		view := model.View()

		assert.Contains(t, view, "💬 Chat Mode")
		assert.Contains(t, view, "🤖 Assistant is thinking")
	})

	t.Run("error state", func(t *testing.T) {
		model.state = ChatStateError
		model.error = "Test error"
		view := model.View()

		assert.Contains(t, view, "💬 Chat Mode - Error")
		assert.Contains(t, view, "❌ Test error")
		assert.Contains(t, view, "Press Esc to continue")
	})

	t.Run("response state with messages", func(t *testing.T) {
		model.state = ChatStateResponse

		// Add some messages
		model.chatMode.conversation.Messages = []Message{
			{Role: "user", Content: "Hello", Timestamp: time.Now()},
			{Role: "assistant", Content: "Hi there!", Timestamp: time.Now()},
		}

		view := model.View()

		assert.Contains(t, view, "💬 Chat Mode")
		assert.Contains(t, view, "You (")
		assert.Contains(t, view, "Assistant (")
		assert.Contains(t, view, "Hello")
		assert.Contains(t, view, "Hi there!")
	})
}

func TestRenderChatHistory(t *testing.T) {
	cfg := &config.Config{Provider: "mock", Model: "test-model", Markdown: true}
	provider := &MockProvider{}
	ui := ui.New()

	chatMode := NewChatMode(cfg, provider, ui)
	model := NewChatModel(chatMode, "")

	t.Run("empty conversation", func(t *testing.T) {
		model.chatMode.conversation.Messages = []Message{}
		history := model.renderChatHistory()

		assert.Contains(t, history, "No messages yet")
	})

	t.Run("conversation with messages", func(t *testing.T) {
		model.chatMode.conversation.Messages = []Message{
			{Role: "user", Content: "Hello", Timestamp: time.Now()},
			{Role: "assistant", Content: "**Hello!** How can I help?", Timestamp: time.Now()},
		}

		history := model.renderChatHistory()

		assert.Contains(t, history, "You (")
		assert.Contains(t, history, "Assistant (")
		assert.Contains(t, history, "Hello")
		// Should contain rendered markdown
		assert.Contains(t, history, "How can I help")
	})

	t.Run("system messages excluded", func(t *testing.T) {
		model.chatMode.conversation.Messages = []Message{
			{Role: "system", Content: "You are helpful", Timestamp: time.Now()},
			{Role: "user", Content: "Hello", Timestamp: time.Now()},
		}

		history := model.renderChatHistory()

		assert.NotContains(t, history, "You are helpful")
		assert.Contains(t, history, "Hello")
	})
}
