package providers

import (
	"github.com/anschmieg/gpt-cli/internal/config"
)

// Provider interface defines the contract for AI providers
type Provider interface {
	CallProvider(prompt string) (string, error)
	StreamProvider(prompt string) (<-chan string, <-chan error)
	GetName() string
}

// NewProvider creates a new provider based on the configuration
func NewProvider(providerName string, config *config.Config) Provider {
	switch providerName {
	case "openai":
		return NewOpenAIProvider(config)
	case "copilot":
		return NewCopilotProvider(config)
	case "gemini":
		return NewGeminiProvider(config)
	default:
		return NewOpenAIProvider(config) // Default fallback
	}
}

// ProviderError represents an error from a provider
type ProviderError struct {
	Message  string
	Code     string
	Original error
}

func (e *ProviderError) Error() string {
	return e.Message
}

// NewProviderError creates a new provider error
func NewProviderError(message, code string, original error) *ProviderError {
	return &ProviderError{
		Message:  message,
		Code:     code,
		Original: original,
	}
}