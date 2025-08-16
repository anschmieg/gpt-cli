package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// CoreConfig represents the configuration for the core processing logic
type CoreConfig struct {
	Provider       string
	Model          string
	Temperature    float64
	System         string
	File           string
	Verbose        bool
	AutoRetryModel bool
	Prompt         string
	UseMarkdown    bool
	Stream         bool
	SuggestMode    bool
}

// ProviderOptions contains options passed to provider adapters
type ProviderOptions struct {
	APIKey  string
	BaseURL string
}

// ProviderResponse represents the response from a provider
type ProviderResponse struct {
	Text     string
	Markdown string
}

// buildProviderOptions creates provider options based on environment variables
func buildProviderOptions(provider string) (*ProviderOptions, error) {
	provider = strings.ToLower(provider)
	opts := &ProviderOptions{}

	switch provider {
	case "openai":
		opts.APIKey = os.Getenv("OPENAI_API_KEY")
		opts.BaseURL = os.Getenv("OPENAI_API_BASE")
	case "copilot":
		opts.APIKey = os.Getenv("COPILOT_API_KEY")
		opts.BaseURL = os.Getenv("COPILOT_API_BASE")
	case "gemini":
		opts.APIKey = os.Getenv("GEMINI_API_KEY")
		// Gemini doesn't use custom base URLs typically
	}

	return opts, nil
}

// getDefaultModel returns the default model for a given provider
func getDefaultModel(provider string) string {
	provider = strings.ToLower(provider)
	switch provider {
	case "gemini":
		return "gemini-2.0-flash"
	default:
		return "gpt-4o-mini" // Global default
	}
}

// runCore is the main orchestration function that handles provider calls and output
func runCore(config *CoreConfig, providerOpts *ProviderOptions) error {
	// Handle suggestion mode separately
	if config.SuggestMode {
		return runSuggestionMode(config, providerOpts)
	}

	// Set default system prompt if not provided
	if config.System == "" {
		config.System = "You are an AI assistant called via CLI. Respond concisely and clearly, focusing only on the user's prompt. Include only very brief explanations unless explicitly asked."
	}

	// Set default model if not provided
	if config.Model == "" {
		config.Model = getDefaultModel(config.Provider)
	}

	if config.Verbose {
		log.Printf("Config: Provider=%s, Model=%s, Temperature=%f, Prompt=%s",
			config.Provider, config.Model, config.Temperature, config.Prompt)
	}

	// Try streaming first if requested
	if config.Stream {
		if err := tryStreamingProvider(config, providerOpts); err == nil {
			return nil
		}
		// Fall back to non-streaming if streaming fails
		if config.Verbose {
			log.Printf("Streaming failed, falling back to non-streaming")
		}
	}

	// Non-streaming path
	response, err := callProvider(config, providerOpts)
	if err != nil {
		// Handle model not supported error with retry
		if config.AutoRetryModel && isModelNotSupportedError(err) {
			if config.Verbose {
				log.Printf("Model not supported, retrying without model...")
			}
			// Retry without explicit model
			retryConfig := *config
			retryConfig.Model = ""
			response, err = callProvider(&retryConfig, providerOpts)
			if err != nil {
				return fmt.Errorf("%v (after retry)", err)
			}
		} else {
			return err
		}
	}

	if config.Verbose {
		log.Printf("Raw response: Text=%d chars, Markdown=%d chars", 
			len(response.Text), len(response.Markdown))
	}

	// Output response based on markdown preference
	if !config.UseMarkdown {
		// Prefer text, fall back to markdown
		if response.Text != "" {
			fmt.Print(response.Text)
		} else {
			// Even for non-markdown preference, render markdown nicely
			renderer := NewMarkdownRenderer(false) // No colors for plain text mode
			fmt.Print(renderer.Render(response.Markdown))
		}
	} else {
		// Prefer markdown, fall back to text
		if response.Markdown != "" {
			// Render markdown with ANSI formatting
			renderer := NewMarkdownRenderer(true) // Enable colors for markdown mode
			fmt.Print(renderer.Render(response.Markdown))
		} else {
			fmt.Print(response.Text)
		}
	}

	return nil
}

// isModelNotSupportedError checks if an error indicates model not supported
func isModelNotSupportedError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "model_not_supported") ||
		strings.Contains(errStr, "model is not supported") ||
		strings.Contains(errStr, "requested model is not supported")
}

// tryStreamingProvider attempts to use streaming API - now implemented in streaming.go

// callProvider calls the appropriate provider adapter
func callProvider(config *CoreConfig, providerOpts *ProviderOptions) (*ProviderResponse, error) {
	provider := strings.ToLower(config.Provider)
	
	switch provider {
	case "openai":
		return callOpenAIProvider(config, providerOpts)
	case "copilot":
		return callCopilotProvider(config, providerOpts)
	case "gemini":
		return callGeminiProvider(config, providerOpts)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", config.Provider)
	}
}