package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/anschmieg/gpt-cli/internal/config"
	"github.com/anschmieg/gpt-cli/internal/modes"
	"github.com/anschmieg/gpt-cli/internal/providers"
	"github.com/anschmieg/gpt-cli/internal/ui"
	"github.com/anschmieg/gpt-cli/internal/utils"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

var (
	provider    string
	model       string
	temperature float64
	verbose     bool
	markdown    bool
	system      string
	stream      bool
	shellMode   bool
	chatMode    bool
)

var rootCmd = &cobra.Command{
	Use:   "gpt-cli [prompt]",
	Short: "A CLI tool for interacting with AI providers",
	Long: `gpt-cli is a command-line interface for interacting with various AI providers
including OpenAI, GitHub Copilot, and Google Gemini.`,
	Run: runCLI,
}

func init() {
	rootCmd.Flags().StringVar(&provider, "provider", "copilot", "API provider (openai, copilot, gemini)")
	rootCmd.Flags().StringVar(&model, "model", "gpt-4o-mini", "Model name")
	rootCmd.Flags().Float64Var(&temperature, "temperature", 0.6, "Temperature (0.0-2.0)")
	rootCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	rootCmd.Flags().BoolVar(&markdown, "markdown", true, "Enable markdown rendering")
	rootCmd.Flags().StringVar(&system, "system", "", "System prompt")
	rootCmd.Flags().BoolVar(&stream, "stream", false, "Enable streaming responses")
	rootCmd.Flags().BoolVar(&shellMode, "shell", false, "Shell suggestion mode - suggest bash commands with safety ratings")
	rootCmd.Flags().BoolVar(&chatMode, "chat", false, "Chat mode - interactive TUI with conversation memory")
}

func runCLI(cmd *cobra.Command, args []string) {
	// Create configuration
	cfg := &config.Config{
		Provider:    provider,
		Model:       model,
		Temperature: temperature,
		Verbose:     verbose,
		Markdown:    markdown,
		System:      system,
	}

	// Override with environment variables if not set via flags
	if !cmd.Flags().Changed("provider") {
		if envProvider := os.Getenv("GPT_CLI_PROVIDER"); envProvider != "" {
			cfg.Provider = envProvider
		}
	}

	if !cmd.Flags().Changed("model") {
		if envModel := os.Getenv("GPT_CLI_MODEL"); envModel != "" {
			cfg.Model = envModel
		}
	}

	if !cmd.Flags().Changed("temperature") {
		if envTemp := os.Getenv("GPT_CLI_TEMPERATURE"); envTemp != "" {
			if temp, err := strconv.ParseFloat(envTemp, 64); err == nil {
				cfg.Temperature = temp
			}
		}
	}

	// Set provider-specific configurations
	switch cfg.Provider {
	case "openai":
		cfg.APIKey = os.Getenv("OPENAI_API_KEY")
		cfg.BaseURL = getEnvOrDefault("OPENAI_API_BASE", "https://api.openai.com")
	case "copilot":
		cfg.APIKey = os.Getenv("COPILOT_API_KEY")
		cfg.BaseURL = os.Getenv("COPILOT_API_BASE")
	case "gemini":
		cfg.APIKey = os.Getenv("GEMINI_API_KEY")
		cfg.BaseURL = getEnvOrDefault("GEMINI_API_BASE", "https://generativelanguage.googleapis.com/v1beta/openai")
	}

	// Set default system prompt if not provided
	if cfg.System == "" {
		cfg.System = config.DefaultSystem
	}

	// Create logger
	logger := utils.NewLogger(cfg.Verbose)

	// Check for conflicting modes
	if shellMode && chatMode {
		fmt.Fprintf(os.Stderr, "Error: Cannot use both --shell and --chat modes simultaneously\n")
		os.Exit(1)
	}

	// If prompt provided as arguments, run in non-interactive mode
	if len(args) > 0 {
		prompt := joinArgs(args)

		// Handle different modes
		if shellMode {
			runShellMode(cfg, prompt, logger)
		} else if chatMode {
			fmt.Fprintf(os.Stderr, "Error: Chat mode requires interactive TUI. Remove arguments to use chat mode.\n")
			os.Exit(1)
		} else {
			runNonInteractive(cfg, prompt, logger)
		}
		return
	}

	// No arguments provided - determine mode
	if shellMode {
		fmt.Fprintf(os.Stderr, "Error: Shell mode requires a prompt argument\n")
		fmt.Fprintf(os.Stderr, "Usage: gpt-cli --shell \"your request for a shell command\"\n")
		os.Exit(1)
	} else if chatMode {
		runChatMode(cfg, logger)
		return
	}

	// Print help and exit if no arguments provided and no mode specified
	fmt.Println("No prompt provided. Use --help for usage information.")
	fmt.Println("Available modes:")
	fmt.Println("  gpt-cli \"your prompt\"        - Simple inline mode")
	fmt.Println("  gpt-cli --shell \"your task\"  - Shell command suggestions")
	fmt.Println("  gpt-cli --chat               - Interactive chat with memory")
	fmt.Println("  go run main.go               - Original TUI mode")
	os.Exit(1)
}

func runNonInteractive(cfg *config.Config, prompt string, logger *utils.Logger) {
	provider := providers.NewProvider(cfg.Provider, cfg)
	runNonInteractiveWithProvider(cfg, prompt, logger, provider, stream)
}

// isTerminalFunc is a hookable function to detect terminals. Tests may override this.
var isTerminalFunc = func(fd uintptr) bool { return isatty.IsTerminal(fd) }

// runNonInteractiveWithProvider runs the non-interactive flow using an explicit provider.
// This is exposed to make testing streaming vs non-streaming and TTY behavior easier.
func runNonInteractiveWithProvider(cfg *config.Config, prompt string, logger *utils.Logger, provider providers.Provider, streamFlag bool) {
	ui := ui.New()

	logger.Debugf("Using provider: %s", cfg.Provider)
	logger.Debugf("Using model: %s", cfg.Model)
	logger.Debugf("Temperature: %.2f", cfg.Temperature)

	if streamFlag {
		// Handle streaming
		contentChan, errorChan := provider.StreamProvider(prompt)

		for {
			select {
			case chunk, ok := <-contentChan:
				if !ok {
					fmt.Println() // New line at end
					return
				}
				fmt.Print(chunk)
			case err, ok := <-errorChan:
				if ok && err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
			}
		}
	}

	// Non-streaming
	response, err := provider.CallProvider(prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Render markdown only when enabled and stdout is a terminal
	if cfg.Markdown && isTerminalFunc(os.Stdout.Fd()) {
		fmt.Println(ui.RenderMarkdown(response))
	} else {
		fmt.Println(response)
	}
}

func runShellMode(cfg *config.Config, prompt string, logger *utils.Logger) {
	provider := providers.NewProvider(cfg.Provider, cfg)
	ui := ui.New()

	logger.Debugf("Using provider: %s", cfg.Provider)
	logger.Debugf("Using model: %s", cfg.Model)
	logger.Debugf("Shell mode prompt: %s", prompt)

	shellMode := modes.NewShellMode(cfg, provider, ui)

	err := shellMode.InteractiveMode(prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error in shell mode: %v\n", err)
		os.Exit(1)
	}
}

func runChatMode(cfg *config.Config, logger *utils.Logger) {
	provider := providers.NewProvider(cfg.Provider, cfg)
	ui := ui.New()

	logger.Debugf("Using provider: %s", cfg.Provider)
	logger.Debugf("Using model: %s", cfg.Model)
	logger.Debugf("Starting chat mode")

	chatMode := modes.NewChatMode(cfg, provider, ui)
	model := modes.NewChatModel(chatMode)

	// Create the BubbleTea program
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running chat mode: %v\n", err)
		os.Exit(1)
	}
}

func joinArgs(args []string) string {
	result := ""
	for i, arg := range args {
		if i > 0 {
			result += " "
		}
		result += arg
	}
	return result
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
