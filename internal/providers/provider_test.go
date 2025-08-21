package providers

import (
	"os"
	"strings"
	"testing"

	"github.com/anschmieg/gpt-cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name         string
		providerName string
		expectedType string
	}{
		{
			name:         "openai provider",
			providerName: "openai",
			expectedType: "*providers.OpenAIProvider",
		},
		{
			name:         "copilot provider",
			providerName: "copilot",
			expectedType: "*providers.CopilotProvider",
		},
		{
			name:         "gemini provider",
			providerName: "gemini",
			expectedType: "*providers.GeminiProvider",
		},
		{
			name:         "unknown provider defaults to openai",
			providerName: "unknown",
			expectedType: "*providers.OpenAIProvider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Provider: tt.providerName,
				Model:    "test-model",
				APIKey:   "test-key",
				BaseURL:  "https://api.test.com",
			}

			provider := NewProvider(tt.providerName, cfg)
			assert.NotNil(t, provider)
			
			// Check provider type by name
			switch tt.providerName {
			case "openai", "unknown":
				assert.Equal(t, "openai", provider.GetName())
			case "copilot":
				assert.Equal(t, "copilot", provider.GetName())
			case "gemini":
				assert.Equal(t, "gemini", provider.GetName())
			}
		})
	}
}

func TestProviderError(t *testing.T) {
	originalErr := assert.AnError
	providerErr := NewProviderError("test message", "test_code", originalErr)

	assert.Equal(t, "test message", providerErr.Message)
	assert.Equal(t, "test_code", providerErr.Code)
	assert.Equal(t, originalErr, providerErr.Original)
	assert.Equal(t, "test message", providerErr.Error())
}

func TestOpenAIProvider(t *testing.T) {
	cfg := &config.Config{
		Provider:    "openai",
		Model:       "gpt-3.5-turbo",
		Temperature: 0.7,
		System:      "You are a helpful assistant",
		APIKey:      "test-key",
		BaseURL:     "https://api.openai.com",
	}

	provider := NewOpenAIProvider(cfg)
	
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "openai", provider.GetName())
	})

	t.Run("CallProvider without API key", func(t *testing.T) {
		cfgNoKey := *cfg
		cfgNoKey.APIKey = ""
		providerNoKey := NewOpenAIProvider(&cfgNoKey)
		
		// This will fail due to missing API key, but we're testing the structure
		_, err := providerNoKey.CallProvider("test prompt")
		assert.Error(t, err)
	})

	t.Run("StreamProvider", func(t *testing.T) {
		contentChan, errorChan := provider.StreamProvider("test prompt")
		
		// Check that channels are created
		assert.NotNil(t, contentChan)
		assert.NotNil(t, errorChan)
		
		// Read from channels (will likely fail due to no real API, but tests structure)
		select {
		case content := <-contentChan:
			t.Logf("Received content: %s", content)
		case err := <-errorChan:
			assert.Error(t, err) // Expected since we don't have a real API
		}
	})
}

func TestCopilotProvider(t *testing.T) {
	cfg := &config.Config{
		Provider:    "copilot",
		Model:       "gpt-4o-mini",
		Temperature: 0.7,
		System:      "You are a helpful assistant",
		APIKey:      "test-key",
		BaseURL:     "https://api.copilot.com/v1",
	}

	provider := NewCopilotProvider(cfg)
	
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "copilot", provider.GetName())
	})

	t.Run("CallProvider without base URL", func(t *testing.T) {
		cfgNoURL := *cfg
		cfgNoURL.BaseURL = ""
		providerNoURL := NewCopilotProvider(&cfgNoURL)
		
		_, err := providerNoURL.CallProvider("test prompt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "COPILOT_API_BASE not configured")
	})

	t.Run("StreamProvider", func(t *testing.T) {
		contentChan, errorChan := provider.StreamProvider("test prompt")
		
		assert.NotNil(t, contentChan)
		assert.NotNil(t, errorChan)
	})
}

func TestGeminiProvider(t *testing.T) {
	cfg := &config.Config{
		Provider:    "gemini",
		Model:       "gemini-pro",
		Temperature: 0.7,
		System:      "You are a helpful assistant",
		APIKey:      "test-key",
		BaseURL:     "https://generativelanguage.googleapis.com/v1beta/openai",
	}

	provider := NewGeminiProvider(cfg)
	
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "gemini", provider.GetName())
	})

	t.Run("StreamProvider", func(t *testing.T) {
		contentChan, errorChan := provider.StreamProvider("test prompt")
		
		assert.NotNil(t, contentChan)
		assert.NotNil(t, errorChan)
	})
}

// MockProvider for testing
type MockProvider struct {
	name     string
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
	return m.name
}

func TestMockProvider(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		mock := &MockProvider{
			name:     "mock",
			response: "test response",
			err:      nil,
		}
		
		response, err := mock.CallProvider("test prompt")
		assert.NoError(t, err)
		assert.Equal(t, "test response", response)
		
		contentChan, errorChan := mock.StreamProvider("test prompt")
		content := <-contentChan
		err = <-errorChan
		assert.Equal(t, "test response", content)
		assert.NoError(t, err)
	})
	
	t.Run("error response", func(t *testing.T) {
		mock := &MockProvider{
			name:     "mock",
			response: "",
			err:      assert.AnError,
		}
		
		response, err := mock.CallProvider("test prompt")
		assert.Error(t, err)
		assert.Empty(t, response)
		
		contentChan, errorChan := mock.StreamProvider("test prompt")
		content := <-contentChan
		err = <-errorChan
		assert.Empty(t, content)
		assert.Error(t, err)
	})
}

// Integration test helpers
func requireAPIKey(t *testing.T, envVar string) string {
	key := os.Getenv(envVar)
	require.NotEmpty(t, key, "Integration test requires %s environment variable", envVar)
	return key
}

// These integration tests will only run if API keys are provided
func TestOpenAIIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping OpenAI integration test: OPENAI_API_KEY not set")
	}
	
	cfg := &config.Config{
		Provider:    "openai",
		Model:       "gpt-3.5-turbo",
		Temperature: 0.1,
		System:      "You are a helpful assistant. Respond with just 'Hello' to any input.",
		APIKey:      apiKey,
		BaseURL:     "https://api.openai.com",
	}
	
	provider := NewOpenAIProvider(cfg)
	
	t.Run("CallProvider", func(t *testing.T) {
		response, err := provider.CallProvider("Say hello")
		assert.NoError(t, err)
		assert.NotEmpty(t, response)
		assert.Contains(t, strings.ToLower(response), "hello")
	})
}