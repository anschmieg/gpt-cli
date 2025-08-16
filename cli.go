package main

import (
	"flag"
	"fmt"
	"strings"
)

// CLIConfig represents the configuration parsed from CLI arguments
type CLIConfig struct {
	Provider      string
	Model         string
	Temperature   float64
	System        string
	File          string
	Verbose       bool
	Markdown      bool
	RetryModel    bool
	Stream        bool
	Help          bool
	Prompt        string
}

// parseArgs parses command line arguments and returns a CLIConfig
func parseArgs(args []string) (*CLIConfig, error) {
	config := &CLIConfig{
		Provider:    "copilot",           // Default provider
		Temperature: 0.6,                // Default temperature
		Verbose:     false,              // Default verbose
		Markdown:    true,               // Default markdown enabled
		RetryModel:  false,              // Default retry disabled
	}

	fs := flag.NewFlagSet("gpt-cli", flag.ContinueOnError)
	fs.StringVar(&config.Provider, "provider", config.Provider, "API provider (openai, gemini, copilot, etc)")
	fs.StringVar(&config.Model, "model", "", "Model name")
	fs.Float64Var(&config.Temperature, "temperature", config.Temperature, "Temperature (float)")
	fs.StringVar(&config.System, "system", "", "System prompt")
	fs.StringVar(&config.File, "file", "", "File to upload")
	fs.BoolVar(&config.Verbose, "verbose", config.Verbose, "Enable verbose logging")
	fs.BoolVar(&config.Markdown, "markdown", config.Markdown, "Enable markdown output")
	fs.BoolVar(&config.RetryModel, "retry-model", config.RetryModel, "Retry with default model if specified model fails")
	fs.BoolVar(&config.Stream, "stream", config.Stream, "Enable streaming output")
	fs.BoolVar(&config.Help, "help", config.Help, "Show help")
	fs.BoolVar(&config.Help, "h", config.Help, "Show help")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// Join remaining arguments as the prompt
	config.Prompt = strings.Join(fs.Args(), " ")

	return config, nil
}

// printHelp prints the help message
func printHelp() {
	fmt.Println(`gpt-cli: Portable GPT API Wrapper

Usage: gpt-cli [options] <prompt>

Options:
  --provider     API provider (openai, gemini, etc)
  --model        Model name
  --temperature  Temperature (float)
  --system       System prompt
  --file         File to upload
  --verbose      Enable verbose logging
  --markdown     Enable markdown output (default: true)
  --retry-model  Retry with default model if specified model fails
  --stream       Enable streaming output
  -h, --help     Show help`)
}

// runCLI is the main CLI entry point
func runCLI(args []string) error {
	config, err := parseArgs(args)
	if err != nil {
		return err
	}

	if config.Help || config.Prompt == "" {
		printHelp()
		return nil
	}

	// Load configuration file if it exists
	fileConfig, _ := LoadConfig()
	ApplyConfigDefaults(config, fileConfig)

	// Create core config and run
	coreConfig := &CoreConfig{
		Provider:       config.Provider,
		Model:          config.Model,
		Temperature:    config.Temperature,
		System:         config.System,
		File:           config.File,
		Verbose:        config.Verbose,
		AutoRetryModel: config.RetryModel,
		Prompt:         config.Prompt,
		UseMarkdown:    config.Markdown,
		Stream:         config.Stream,
	}

	// Build provider options from environment
	providerOpts, err := buildProviderOptions(config.Provider)
	if err != nil {
		return err
	}

	// Run the core logic
	return runCore(coreConfig, providerOpts)
}