// Centralized default configuration for the CLI.
export const DEFAULTS = {
  provider: "copilot",
  model: "gpt-4o-mini",
  temperature: 0.6,
  verbose: false,
  markdown: true,
} as const;

export const DEFAULT_MODEL = DEFAULTS.model;
