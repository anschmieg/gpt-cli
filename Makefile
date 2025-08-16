# Makefile for gpt-cli Go rewrite

.PHONY: build test clean mock-server help

# Build the main CLI binary
build:
	go build -o gpt-cli .

# Run all tests
test:
	go test -v ./...

# Run tests with integration tests
test-integration: build
	@echo "Starting mock server..."
	@GPT_CLI_TEST=1 MOCK_SERVER_PORT=8086 go run mock_server.go &
	@sleep 1
	@echo "Running integration tests..."
	@GPT_CLI_TEST=1 MOCK_SERVER_URL=http://127.0.0.1:8086 go test -v -run TestIntegration || true
	@echo "Stopping mock server..."
	@pkill -f "go run mock_server.go" || true

# Start mock server for manual testing
mock-server:
	@echo "Starting mock server on port 8086..."
	@GPT_CLI_TEST=1 go run mock_server.go

# Clean build artifacts
clean:
	rm -f gpt-cli

# Test the CLI manually (requires mock server running)
test-cli: build
	@echo "Testing CLI with mock server..."
	@GPT_CLI_TEST=1 MOCK_SERVER_URL=http://127.0.0.1:8086 ./gpt-cli --verbose "Hello world test"

# Show help
help:
	@echo "Available targets:"
	@echo "  build           - Build the CLI binary"
	@echo "  test            - Run unit tests"
	@echo "  test-integration- Run integration tests with mock server"
	@echo "  mock-server     - Start mock server for manual testing"
	@echo "  test-cli        - Test CLI with mock server (requires mock-server running)"
	@echo "  clean           - Clean build artifacts"
	@echo "  help            - Show this help"

# Default target
all: build