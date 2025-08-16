import { makeOpenAICompatibleAdapter } from "./factory.ts";

const adapter = makeOpenAICompatibleAdapter({
  apiKeyEnv: "OPENAI_API_KEY",
  baseEnv: "OPENAI_API_BASE",
  defaultBase: "https://api.openai.com",
});

export const callProvider = adapter.callProvider;
export const chatCompletionStream = adapter.chatCompletionStream;

// intentionally no local type re-exports; types live in `src/providers/types.ts`
