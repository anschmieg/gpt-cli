import { makeOpenAICompatibleAdapter } from "./factory.ts";

const adapter = makeOpenAICompatibleAdapter({
  apiKeyEnv: "GEMINI_API_KEY",
  baseEnv: undefined,
  defaultBase: "https://generativelanguage.googleapis.com/v1beta/openai",
  // The Google Generative Language OpenAI-compatible base should be
  // converted into the full OpenAI-style chat completions endpoint.
  // Per docs a valid endpoint looks like:
  // https://generativelanguage.googleapis.com/v1beta/openai/chat/completions
  constructUrl: (base) => `${base.replace(/\/$/, "")}/chat/completions`,
});

export const callProvider = adapter.callProvider;
export const chatCompletionStream = adapter.chatCompletionStream;
