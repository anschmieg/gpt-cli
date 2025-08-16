import { makeOpenAICompatibleAdapter } from "./factory.ts";

const adapter = makeOpenAICompatibleAdapter({
  apiKeyEnv: "GEMINI_API_KEY",
  baseEnv: undefined,
  defaultBase: "https://generativelanguage.googleapis.com/v1beta/openai",
  constructUrl: (base) => base.replace(/\/$/, ""),
});

export const callProvider = adapter.callProvider;
export const chatCompletionStream = adapter.chatCompletionStream;
