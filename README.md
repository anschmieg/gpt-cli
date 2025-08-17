# GPT CLI - Go BubbleTea Version

A modern, interactive CLI tool for communicating with various AI providers (OpenAI, GitHub Copilot, Google Gemini) built with Go and BubbleTea.

## Features

- **Interactive TUI**: Beautiful terminal user interface powered by BubbleTea
- **Multiple Providers**: Support for OpenAI, GitHub Copilot, and Google Gemini
- **Markdown Rendering**: Rich text formatting in the terminal
- **Streaming Support**: Real-time response streaming (simulated)
- **Configuration**: Environment variable and flag-based configuration
- **Non-Interactive Mode**: Command-line usage for scripts and automation

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd bubbletea-app

# Install dependencies
go mod tidy

# Build the application
go build -o gpt-cli main.go

# Or run directly
go run main.go
```

## Usage

### Interactive Mode

Run the application without arguments to enter interactive mode:

```bash
go run main.go
```

This will launch a beautiful TUI where you can:
- Type your prompts
- See real-time responses
- Navigate between input and response views
- Quit with `Ctrl+C` or `q`

### Non-Interactive Mode

Use the CLI command for direct queries:

```bash
go run cmd/cli/main.go "What is the capital of France?"
```

### Command Line Options

```bash
go run cmd/cli/main.go --help
```

Available flags:
- `--provider`: API provider (openai, copilot, gemini) [default: copilot]
- `--model`: Model name [default: gpt-4o-mini]
- `--temperature`: Temperature (0.0-2.0) [default: 0.6]
- `--verbose`: Enable verbose logging
- `--markdown`: Enable markdown rendering [default: true]
- `--system`: Custom system prompt
- `--stream`: Enable streaming responses

## Configuration

### Environment Variables

Set up your API keys and configuration:

```bash
# OpenAI
export OPENAI_API_KEY="your-openai-key"
export OPENAI_API_BASE="https://api.openai.com"  # optional

# GitHub Copilot
export COPILOT_API_KEY="your-copilot-key"
export COPILOT_API_BASE="your-copilot-endpoint"

# Google Gemini
export GEMINI_API_KEY="your-gemini-key"
export GEMINI_API_BASE="https://generativelanguage.googleapis.com/v1beta/openai"  # optional

# General settings
export GPT_CLI_PROVIDER="copilot"
export GPT_CLI_MODEL="gpt-4o-mini"
export GPT_CLI_TEMPERATURE="0.6"
export GPT_CLI_VERBOSE="false"
export GPT_CLI_MARKDOWN="true"
```

## Project Structure

```
├── main.go                    # Interactive TUI entry point
├── cmd/cli/main.go           # Non-interactive CLI entry point
├── internal/
│   ├── app/
│   │   └── model.go          # BubbleTea model and application logic
│   ├── config/
│   │   └── config.go         # Configuration management
│   ├── providers/
│   │   ├── provider.go       # Provider interface and factory
│   │   ├── openai.go         # OpenAI provider implementation
│   │   ├── copilot.go        # GitHub Copilot provider implementation
│   │   └── gemini.go         # Google Gemini provider implementation
│   ├── ui/
│   │   └── ui.go             # UI styling and markdown rendering
│   └── utils/
│       └── logger.go         # Logging utilities
├── go.mod
└── README.md
```

## Architecture

### BubbleTea Model

The application uses the BubbleTea framework for the interactive TUI:

- **Model**: Manages application state (input, loading, response, error)
- **Update**: Handles user input and state transitions
- **View**: Renders the current state to the terminal

### Provider System

Modular provider system supporting multiple AI services:

- **Provider Interface**: Common contract for all providers
- **Factory Pattern**: Dynamic provider creation based on configuration
- **Error Handling**: Normalized error responses across providers

### UI System

Rich terminal interface with:

- **Lipgloss Styling**: Beautiful, consistent visual design
- **Markdown Rendering**: Basic markdown support with syntax highlighting
- **Responsive Layout**: Adapts to terminal size

## Development

### Adding a New Provider

1. Create a new file in `internal/providers/`
2. Implement the `Provider` interface
3. Add the provider to the factory in `provider.go`
4. Update configuration handling

### Extending Markdown Support

The markdown renderer in `internal/ui/ui.go` supports:
- Headers (H1-H6)
- Bold and italic text
- Inline code
- Code blocks
- Lists

To add new features, extend the `RenderMarkdown` method.

### Testing

```bash
# Run tests
go test ./...

# Run with verbose output
go test -v ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [BubbleTea](https://github.com/charmbracelet/bubbletea) - The TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- Original TypeScript/Deno implementation for inspiration