package app

import (
	"fmt"
	"strings"

	"github.com/anschmieg/gpt-cli/internal/config"
	"github.com/anschmieg/gpt-cli/internal/providers"
	"github.com/anschmieg/gpt-cli/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the main application state
type Model struct {
	config   *config.Config
	state    AppState
	input    string
	response string
	loading  bool
	error    string
	cursor   int
	width    int
	height   int
	provider providers.Provider
	ui       *ui.UI
}

// AppState represents the current state of the application
type AppState int

const (
	StateInput AppState = iota
	StateLoading
	StateResponse
	StateError
)

// NewModel creates a new application model
func NewModel() *Model {
	cfg := config.NewConfig()
	provider := providers.NewProvider(cfg.Provider, cfg)

	return &Model{
		config:   cfg,
		state:    StateInput,
		provider: provider,
		ui:       ui.New(),
	}
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case ResponseMsg:
		m.state = StateResponse
		m.response = msg.Content
		m.loading = false
		return m, nil

	case ErrorMsg:
		m.state = StateError
		m.error = msg.Error
		m.loading = false
		return m, nil

	case StreamChunkMsg:
		m.response += msg.Chunk
		return m, nil
	}

	return m, nil
}

// handleKeyMsg handles keyboard input
func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		if m.state == StateInput || m.state == StateResponse || m.state == StateError {
			return m, tea.Quit
		}

	case "esc":
		if m.state == StateResponse || m.state == StateError {
			m.state = StateInput
			m.response = ""
			m.error = ""
			return m, nil
		}

	case "enter":
		if m.state == StateInput && strings.TrimSpace(m.input) != "" {
			return m.sendRequest()
		}

	case "backspace":
		if m.state == StateInput && len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}

	default:
		if m.state == StateInput {
			m.input += msg.String()
		}
	}

	return m, nil
}

// sendRequest sends a request to the provider
func (m *Model) sendRequest() (tea.Model, tea.Cmd) {
	m.state = StateLoading
	m.loading = true
	m.response = ""
	m.error = ""

	prompt := strings.TrimSpace(m.input)
	m.input = ""

	return m, m.makeRequest(prompt)
}

// makeRequest creates a command to make a request to the provider
func (m *Model) makeRequest(prompt string) tea.Cmd {
	return func() tea.Msg {
		response, err := m.provider.CallProvider(prompt)
		if err != nil {
			return ErrorMsg{Error: err.Error()}
		}
		return ResponseMsg{Content: response}
	}
}

// View renders the current view
func (m *Model) View() string {
	switch m.state {
	case StateInput:
		return m.renderInputView()
	case StateLoading:
		return m.renderLoadingView()
	case StateResponse:
		return m.renderResponseView()
	case StateError:
		return m.renderErrorView()
	default:
		return "Unknown state"
	}
}

// renderInputView renders the input view
func (m *Model) renderInputView() string {
	title := m.ui.TitleStyle.Render("GPT CLI")

	providerInfo := m.ui.SubtitleStyle.Render(fmt.Sprintf(
		"Provider: %s | Model: %s | Temperature: %.1f",
		m.config.Provider,
		m.config.Model,
		m.config.Temperature,
	))

	prompt := m.ui.PromptStyle.Render("Enter your prompt:")
	input := m.ui.InputStyle.Render("> " + m.input + "█")

	help := m.ui.HelpStyle.Render("Press Enter to send • Ctrl+C or q to quit")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		providerInfo,
		"",
		prompt,
		input,
		"",
		help,
	)

	return m.ui.ContainerStyle.Render(content)
}

// renderLoadingView renders the loading view
func (m *Model) renderLoadingView() string {
	title := m.ui.TitleStyle.Render("GPT CLI")
	loading := m.ui.LoadingStyle.Render("🤖 Thinking...")
	help := m.ui.HelpStyle.Render("Press Ctrl+C to cancel")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		loading,
		"",
		help,
	)

	return m.ui.ContainerStyle.Render(content)
}

// renderResponseView renders the response view
func (m *Model) renderResponseView() string {
	title := m.ui.TitleStyle.Render("GPT CLI - Response")

	// Render markdown if enabled
	var response string
	if m.config.Markdown {
		// Prefer the fragment-aware renderer supplied by UI.New(). It will
		// render appropriately for TTY vs non-TTY and handle fragment-safe
		// normalization.
		if m.ui.Renderer != nil {
			response = m.ui.Renderer.Render(m.response)
		} else {
			response = m.response
		}
	} else {
		response = m.response
	}

	responseBox := m.ui.ResponseStyle.Render(response)
	help := m.ui.HelpStyle.Render("Press Esc for new prompt • Ctrl+C or q to quit")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		responseBox,
		"",
		help,
	)

	return m.ui.ContainerStyle.Render(content)
}

// renderErrorView renders the error view
func (m *Model) renderErrorView() string {
	title := m.ui.TitleStyle.Render("GPT CLI - Error")
	error := m.ui.ErrorStyle.Render("❌ " + m.error)
	help := m.ui.HelpStyle.Render("Press Esc to try again • Ctrl+C or q to quit")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		error,
		"",
		help,
	)

	return m.ui.ContainerStyle.Render(content)
}

// Message types
type ResponseMsg struct {
	Content string
}

type ErrorMsg struct {
	Error string
}

type StreamChunkMsg struct {
	Chunk string
}

// State returns the current state (for testing)
func (m *Model) State() AppState {
	return m.state
}
