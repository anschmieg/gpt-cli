package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// MockProvider for testing
type MockProvider struct {
	response string
	err      error
}

func (m *MockProvider) CallProvider(prompt string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func (m *MockProvider) StreamProvider(prompt string) (<-chan string, <-chan error) {
	contentChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	if m.err != nil {
		errorChan <- m.err
	} else {
		contentChan <- m.response
	}

	close(contentChan)
	close(errorChan)

	return contentChan, errorChan
}

func (m *MockProvider) GetName() string {
	return "mock"
}

func TestNewModel(t *testing.T) {
	model := NewModel()

	assert.NotNil(t, model)
	assert.NotNil(t, model.config)
	assert.NotNil(t, model.provider)
	assert.NotNil(t, model.ui)
	assert.Equal(t, StateInput, model.state)
	assert.Empty(t, model.input)
	assert.Empty(t, model.response)
	assert.False(t, model.loading)
	assert.Empty(t, model.error)
}

func TestModelInit(t *testing.T) {
	model := NewModel()
	cmd := model.Init()

	assert.NotNil(t, cmd) // Should return a command
}

func TestModelUpdate(t *testing.T) {
	model := NewModel()

	t.Run("window size message", func(t *testing.T) {
		msg := tea.WindowSizeMsg{Width: 80, Height: 24}
		updatedModel, cmd := model.Update(msg)

		assert.Nil(t, cmd)
		appModel := updatedModel.(*Model)
		assert.Equal(t, 80, appModel.width)
		assert.Equal(t, 24, appModel.height)
	})

	t.Run("response message", func(t *testing.T) {
		msg := ResponseMsg{Content: "Test response"}
		updatedModel, cmd := model.Update(msg)

		assert.Nil(t, cmd)
		appModel := updatedModel.(*Model)
		assert.Equal(t, StateResponse, appModel.state)
		assert.Equal(t, "Test response", appModel.response)
		assert.False(t, appModel.loading)
	})

	t.Run("error message", func(t *testing.T) {
		msg := ErrorMsg{Error: "Test error"}
		updatedModel, cmd := model.Update(msg)

		assert.Nil(t, cmd)
		appModel := updatedModel.(*Model)
		assert.Equal(t, StateError, appModel.state)
		assert.Equal(t, "Test error", appModel.error)
		assert.False(t, appModel.loading)
	})

	t.Run("stream chunk message", func(t *testing.T) {
		// Start with empty response
		model.response = ""
		msg := StreamChunkMsg{Chunk: "chunk"}
		updatedModel, cmd := model.Update(msg)

		assert.Nil(t, cmd)
		appModel := updatedModel.(*Model)
		assert.Equal(t, "chunk", appModel.response)
	})
}

func TestModelKeyHandling(t *testing.T) {
	model := NewModel()
	// Replace provider with mock for testing
	model.provider = &MockProvider{response: "Test response"}

	t.Run("typing input", func(t *testing.T) {
		model.state = StateInput
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}
		updatedModel, cmd := model.Update(msg)

		assert.Nil(t, cmd)
		appModel := updatedModel.(*Model)
		assert.Equal(t, "h", appModel.input)
	})

	t.Run("backspace", func(t *testing.T) {
		model.state = StateInput
		model.input = "hello"
		msg := tea.KeyMsg{Type: tea.KeyBackspace}
		updatedModel, cmd := model.Update(msg)

		assert.Nil(t, cmd)
		appModel := updatedModel.(*Model)
		assert.Equal(t, "hell", appModel.input)
	})

	t.Run("enter with input", func(t *testing.T) {
		model.state = StateInput
		model.input = "test prompt"
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, cmd := model.Update(msg)

		assert.NotNil(t, cmd)
		appModel := updatedModel.(*Model)
		assert.Equal(t, StateLoading, appModel.state)
		assert.True(t, appModel.loading)
		assert.Empty(t, appModel.input) // Should be cleared
	})

	t.Run("quit", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyCtrlC}
		updatedModel, _ := model.Update(msg)

		// Just check that the model is returned
		assert.NotNil(t, updatedModel)
	})

	t.Run("escape from response state", func(t *testing.T) {
		model.state = StateResponse
		model.response = "Some response"

		msg := tea.KeyMsg{Type: tea.KeyEsc}
		updatedModel, cmd := model.Update(msg)

		assert.Nil(t, cmd)
		appModel := updatedModel.(*Model)
		assert.Equal(t, StateInput, appModel.state)
		assert.Empty(t, appModel.response)
	})
}

func TestModelView(t *testing.T) {
	model := NewModel()
	model.width = 80
	model.height = 24

	t.Run("input state", func(t *testing.T) {
		model.state = StateInput
		view := model.View()

		assert.Contains(t, view, "GPT CLI")
		assert.Contains(t, view, "Enter your prompt")
		assert.Contains(t, view, "Press Enter to send")
	})

	t.Run("loading state", func(t *testing.T) {
		model.state = StateLoading
		model.loading = true
		view := model.View()

		assert.Contains(t, view, "GPT CLI")
		assert.Contains(t, view, "🤖 Thinking")
	})

	t.Run("response state", func(t *testing.T) {
		model.state = StateResponse
		model.response = "Test response"
		view := model.View()

		assert.Contains(t, view, "GPT CLI - Response")
		assert.Contains(t, view, "Test response")
		assert.Contains(t, view, "Press Esc for new prompt")
	})

	t.Run("error state", func(t *testing.T) {
		model.state = StateError
		model.error = "Test error"
		view := model.View()

		assert.Contains(t, view, "GPT CLI - Error")
		assert.Contains(t, view, "❌ Test error")
		assert.Contains(t, view, "Press Esc to try again")
	})
}

func TestMakeRequest(t *testing.T) {
	model := NewModel()
	mockProvider := &MockProvider{response: "Mock response"}
	model.provider = mockProvider

	cmd := model.makeRequest("test prompt")
	assert.NotNil(t, cmd)

	// Execute the command to test it
	msg := cmd()
	responseMsg, ok := msg.(ResponseMsg)
	assert.True(t, ok)
	assert.Equal(t, "Mock response", responseMsg.Content)
}

func TestMakeRequestWithError(t *testing.T) {
	model := NewModel()
	mockProvider := &MockProvider{err: assert.AnError}
	model.provider = mockProvider

	cmd := model.makeRequest("test prompt")
	assert.NotNil(t, cmd)

	// Execute the command to test it
	msg := cmd()
	errorMsg, ok := msg.(ErrorMsg)
	assert.True(t, ok)
	assert.Equal(t, assert.AnError.Error(), errorMsg.Error)
}

func TestSendRequest(t *testing.T) {
	model := NewModel()
	model.input = "   test prompt   "
	model.state = StateInput

	updatedModel, cmd := model.sendRequest()

	assert.NotNil(t, cmd)
	appModel := updatedModel.(*Model)
	assert.Equal(t, StateLoading, appModel.state)
	assert.True(t, appModel.loading)
	assert.Empty(t, appModel.input) // Should be trimmed and cleared
	assert.Empty(t, appModel.response)
	assert.Empty(t, appModel.error)
}
