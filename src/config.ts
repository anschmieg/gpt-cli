// Centralized default configuration for the CLI.
export const DEFAULTS = {
  provider: "copilot",
  model: "gpt-4o-mini",
  temperature: 0.6,
  verbose: false,
  markdown: true,
} as const;

// Local mock server used by integration tests
export const MOCK_SERVER_URL = "http://127.0.0.1:8086";
