package main

import (
    "os"
    "testing"
)

// TestMain ensures provider-related environment variables are cleared so
// unit tests never make real network calls or depend on user env.
func TestMain(m *testing.M) {
    // Clear provider env vars that could flip code paths to real HTTP
    _ = os.Unsetenv("OPENAI_API_KEY")
    _ = os.Unsetenv("OPENAI_API_BASE")
    _ = os.Unsetenv("COPILOT_API_KEY")
    _ = os.Unsetenv("COPILOT_API_BASE")
    _ = os.Unsetenv("GEMINI_API_KEY")
    _ = os.Unsetenv("GEMINI_API_BASE")
    _ = os.Unsetenv("GPT_CLI_PROVIDER")

    os.Exit(m.Run())
}

