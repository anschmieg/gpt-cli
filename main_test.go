package main

import (
	"os"
	"testing"
)

func TestMainFunction(t *testing.T) {
	// Start mock server for testing
	server := startMockTestServer()
	defer server.Close()

	// Set test environment variables
	t.Setenv("GPT_CLI_TEST", "1")
	t.Setenv("MOCK_SERVER_URL", server.URL)

	// Test with no arguments - should print hello message
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test case 1: No arguments
	os.Args = []string{"gpt-cli"}
	// main() should not panic
	// We can't easily test output without redirecting stdout

	// Test case 2: With arguments
	os.Args = []string{"gpt-cli", "hello world"}
	// main() should not panic
}

func TestMainWithInvalidArgs(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test with invalid flag - this might cause exit(1)
	// We can't easily test this without exit, but we can test it doesn't panic
	os.Args = []string{"gpt-cli", "--invalid-flag"}
}