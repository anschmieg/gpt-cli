# Makefile for gpt-cli (Go rewrite)

.PHONY: build test clean coverage help

# Build the main CLI binary
build:
	go build -o gpt-cli .

# Run all unit tests (fast, no external services)
test:
	go test -v ./...

# Generate coverage report and HTML
.PHONY: coverage
coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f gpt-cli coverage.out coverage.html

# Show help
help:
	@echo "Available targets:"
	@echo "  build       - Build the CLI binary"
	@echo "  test        - Run unit tests"
	@echo "  coverage    - Generate coverage.html"
	@echo "  clean       - Remove build artifacts and coverage files"
	@echo "  help        - Show this help"
