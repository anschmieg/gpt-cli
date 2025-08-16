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

// Per-provider default model overrides. Add entries here when a provider
// expects a different default model than the global DEFAULTS.model.
export const DEFAULT_MODEL_BY_PROVIDER: Record<string, string> = {
  // Google Gemini/Generative Language adapter default
  // Use a Gemini-specific model name by default per provider docs.
  // Users can still override with `--model`.
  gemini: "gemini-2.0-flash",
};
