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

// NewProviderHook allows tests (e.g., e2e) to override provider construction.
// If set, NewProvider will delegate to this hook.
var NewProviderHook func(providerName string, cfg *config.Config) Provider

// NewProvider creates a new provider based on the configuration
func NewProvider(providerName string, config *config.Config) Provider {
    if NewProviderHook != nil {
        if p := NewProviderHook(providerName, config); p != nil {
            return p
        }
    }
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
