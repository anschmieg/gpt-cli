# gpt-cli (Go Version)

![CI Status](https://github.com/anschmieg/gpt-cli/actions/workflows/ci.yml/badge.svg)

A portable Go CLI that acts as a thin wrapper around provider APIs (OpenAI-compatible by default) and contains a fast local mock server for testing.

**This is a complete rewrite of the original Deno CLI in Go while maintaining 100% functional compatibility.**

## What the program does

- `main.go` is the CLI entry point that parses command-line arguments and orchestrates provider calls
- `core.go` contains the core logic that handles provider interactions, retries, and response rendering
- `providers.go` implements the provider adapters for OpenAI, Copilot, and Gemini APIs
- `mock_server.go` provides a local mock OpenAI-compatible server for testing

## Features

- ✅ Multiple provider support (OpenAI, GitHub Copilot, Google Gemini)
- ✅ Configurable models, temperature, and system prompts
- ✅ Automatic model retry on failures
- ✅ Markdown output support
- ✅ Verbose logging
- ✅ Environment-based configuration
- ✅ Comprehensive testing with mock server
- 🚧 Streaming support (planned)

## Building and Usage

### Build
```bash
make build
# or
go build -o gpt-cli .
```

### Basic Usage
```bash
./gpt-cli "Hello, world!"
./gpt-cli --provider openai --model gpt-4 "Write a haiku"
./gpt-cli --verbose --temperature 0.8 "Explain quantum computing"
```

### Options
```
--provider     API provider (openai, gemini, copilot) [default: copilot]
--model        Model name [default: gpt-4o-mini, gemini-2.0-flash for Gemini]
--temperature  Temperature (0.0-2.0) [default: 0.6]
--system       System prompt
--file         File to upload (placeholder)
--verbose      Enable verbose logging
--markdown     Enable markdown output [default: true]
--retry-model  Retry with default model if specified model fails
--stream       Enable streaming output (not yet implemented)
-h, --help     Show help
```

## Environment Variables

- `OPENAI_API_KEY` - OpenAI API key
- `OPENAI_API_BASE` - Custom OpenAI API base URL
- `COPILOT_API_KEY` - GitHub Copilot API key  
- `COPILOT_API_BASE` - Custom Copilot API base URL
- `GEMINI_API_KEY` - Google Gemini API key
- `GPT_CLI_TEST` - Set to "1" to enable test mode (uses mock server)
- `MOCK_SERVER_URL` - Override mock server URL for testing

## Testing

### Unit Tests
```bash
make test
# or
go test -v ./...
```

### Integration Tests
```bash
make test-integration
```

### Manual Testing with Mock Server
```bash
# Terminal 1: Start mock server
make mock-server

# Terminal 2: Test CLI
make test-cli
```

## Default Configuration

- **Provider**: `copilot` (GitHub Copilot)
- **Model**: `gpt-4o-mini` (global default), `gemini-2.0-flash` (for Gemini)
- **Temperature**: `0.6`
- **Markdown**: `true`
- **System Prompt**: "You are an AI assistant called via CLI. Respond concisely and clearly, focusing only on the user's prompt. Include only very brief explanations unless explicitly asked."

## Differences from Deno Version

This Go rewrite maintains 100% functional compatibility with the original Deno version while:

- Using Go's standard library instead of Deno runtime
- Implementing the same CLI interface and behavior
- Preserving all provider adapters and error handling
- Maintaining the same testing approach with mock servers
- Supporting the same environment variables and configuration

## License

This project is MIT-licensed.