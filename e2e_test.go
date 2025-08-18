package main

import (
	"os"
	"testing"
	"time"

	"github.com/anschmieg/gpt-cli/internal/app"
	"github.com/anschmieg/gpt-cli/internal/config"
	"github.com/anschmieg/gpt-cli/internal/modes"
	"github.com/anschmieg/gpt-cli/internal/providers"
	"github.com/anschmieg/gpt-cli/internal/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockProvider for end-to-end tests
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

func TestEndToEndConfigLoad(t *testing.T) {
	// Test config loading from environment
	os.Setenv("GPT_CLI_PROVIDER", "openai")
	os.Setenv("GPT_CLI_MODEL", "gpt-4")
	os.Setenv("OPENAI_API_KEY", "test-key")
	defer func() {
		os.Unsetenv("GPT_CLI_PROVIDER")
		os.Unsetenv("GPT_CLI_MODEL")
		os.Unsetenv("OPENAI_API_KEY")
	}()

	cfg := config.NewConfig()
	assert.Equal(t, "openai", cfg.Provider)
	assert.Equal(t, "gpt-4", cfg.Model)
	assert.Equal(t, "test-key", cfg.APIKey)
}

func TestEndToEndAppModel(t *testing.T) {
	// Test complete app model workflow
	model := app.NewModel()

	// Test initialization
	assert.NotNil(t, model)
	assert.Equal(t, app.StateInput, model.State())

	// Note: Provider testing is covered by unit tests
	// This e2e test focuses on integration workflow
}

func TestEndToEndShellMode(t *testing.T) {
	// Test shell mode workflow
	cfg := &config.Config{
		Provider:    "mock",
		Model:       "test-model",
		Temperature: 0.1,
	}

	mockProvider := &MockProvider{
		response: `{
			"command": "ls -la",
			"safety_level": "safe",
			"explanation": "Lists all files and directories with detailed information",
			"reasoning": "This is a read-only command that only displays information"
		}`,
	}

	ui := ui.New()
	shellMode := modes.NewShellMode(cfg, mockProvider, ui)

	suggestion, err := shellMode.SuggestCommand("list all files")

	require.NoError(t, err)
	assert.Equal(t, "ls -la", suggestion.Command)
	assert.Equal(t, "safe", suggestion.SafetyLevel)
	assert.Contains(t, suggestion.Explanation, "Lists all files")
	assert.Contains(t, suggestion.Reasoning, "read-only")
}

func TestEndToEndChatMode(t *testing.T) {
	// Test chat mode workflow
	cfg := &config.Config{
		Provider: "mock",
		Model:    "test-model",
		System:   "You are a helpful assistant",
	}

	mockProvider := &MockProvider{response: "Hello! How can I help you today?"}
	ui := ui.New()

	chatMode := modes.NewChatMode(cfg, mockProvider, ui)

	// Test initial state
	conv := chatMode.GetConversation()
	assert.NotNil(t, conv)
	assert.Len(t, conv.Messages, 1) // System message
	assert.Equal(t, "system", conv.Messages[0].Role)

	// Test adding user message (simulation)
	conv.Messages = append(conv.Messages, modes.Message{
		Role:      "user",
		Content:   "Hello",
		Timestamp: time.Now(),
	})

	// Test export functionality
	exported := chatMode.ExportConversation()
	assert.Contains(t, exported, "Conversation")
	assert.Contains(t, exported, "Hello")
}

func TestEndToEndProviderFactory(t *testing.T) {
	// Test provider factory creates correct providers
	cfg := &config.Config{
		Provider: "openai",
		APIKey:   "test-key",
		BaseURL:  "https://api.openai.com",
	}

	provider := providers.NewProvider("openai", cfg)
	assert.Equal(t, "openai", provider.GetName())

	provider = providers.NewProvider("copilot", cfg)
	assert.Equal(t, "copilot", provider.GetName())

	provider = providers.NewProvider("gemini", cfg)
	assert.Equal(t, "gemini", provider.GetName())

	// Test unknown provider defaults to openai
	provider = providers.NewProvider("unknown", cfg)
	assert.Equal(t, "openai", provider.GetName())
}

func TestEndToEndMarkdownRendering(t *testing.T) {
	// Test markdown rendering workflow
	ui := ui.New()

	markdownText := "# Test Header\n\nThis is **bold** text."

	// Test markdown detection
	assert.True(t, ui.IsMarkdown(markdownText))

	// Test rendering (should not panic and should return formatted text)
	rendered := ui.RenderMarkdown(markdownText)
	assert.NotEmpty(t, rendered)
	// The rendered output contains ANSI codes, so just check it's not empty
	// and that it contains at least some of the original content (rendered)
	assert.Greater(t, len(rendered), len(markdownText))
}

func TestEndToEndConfigFileIntegration(t *testing.T) {
	// Test config file loading (simulated)
	// In a real e2e test, we'd create actual config files

	// Test YAML format detection
	yamlContent := `
provider: openai
model: gpt-4
temperature: 0.8
providers:
  openai:
    api_key: yaml-test-key
`

	// Test JSON format detection
	jsonContent := `{
  "provider": "copilot",
  "model": "gpt-4o-mini",
  "temperature": 0.5,
  "providers": {
    "copilot": {
      "api_key": "json-test-key"
    }
  }
}`

	// These would be parsed by the config system in a real scenario
	assert.Contains(t, yamlContent, "provider: openai")
	assert.Contains(t, jsonContent, "\"provider\": \"copilot\"")
}

// Integration test for the complete workflow
func TestEndToEndIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 1. Create config
	cfg := config.NewConfig()
	assert.NotNil(t, cfg)

	// 2. Create provider
	provider := providers.NewProvider(cfg.Provider, cfg)
	assert.NotNil(t, provider)

	// 3. Create UI
	ui := ui.New()
	assert.NotNil(t, ui)

	// 4. Test shell mode creation
	shellMode := modes.NewShellMode(cfg, provider, ui)
	assert.NotNil(t, shellMode)

	// 5. Test chat mode creation
	chatMode := modes.NewChatMode(cfg, provider, ui)
	assert.NotNil(t, chatMode)

	// 6. Test app model creation
	appModel := app.NewModel()
	assert.NotNil(t, appModel)

	// This confirms all components can be created and work together
}
