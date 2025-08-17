package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"bubbletea-app/internal/config"
	"bubbletea-app/internal/providers"
	"bubbletea-app/internal/ui"
	"bubbletea-app/internal/utils"
)

var (
	provider    string
	model       string
	temperature float64
	verbose     bool
	markdown    bool
	system      string
	stream      bool
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

	// If prompt provided as arguments, run in non-interactive mode
	if len(args) > 0 {
		prompt := joinArgs(args)
		runNonInteractive(cfg, prompt, logger)
		return
	}

	// Print help and exit if no arguments provided
	fmt.Println("No prompt provided. Use --help for usage information.")
	fmt.Println("For interactive mode, run the main application:")
	fmt.Println("  go run main.go")
	os.Exit(1)
}

func runNonInteractive(cfg *config.Config, prompt string, logger *utils.Logger) {
	provider := providers.NewProvider(cfg.Provider, cfg)
	ui := ui.New()

	logger.Debugf("Using provider: %s", cfg.Provider)
	logger.Debugf("Using model: %s", cfg.Model)
	logger.Debugf("Temperature: %.2f", cfg.Temperature)

	if stream {
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
	} else {
		// Handle non-streaming
		response, err := provider.CallProvider(prompt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if cfg.Markdown {
			fmt.Println(ui.RenderMarkdown(response))
		} else {
			fmt.Println(response)
		}
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