import { makeOpenAICompatibleAdapter } from "./factory.ts";

// Copilot exposes an OpenAI-compatible endpoint; the factory will normalize
// url construction for us via defaultBase detection. We still read COPILOT_API_BASE
// and COPILOT_API_KEY by env name for convenience.
const adapter = makeOpenAICompatibleAdapter({
  apiKeyEnv: "COPILOT_API_KEY",
  baseEnv: "COPILOT_API_BASE",
  defaultBase: "",
  constructUrl: (base) => {
    const trimmed = base.replace(/\/$/, "");
    if (/\/v1\/chat\/completions$/.test(trimmed)) return trimmed;
    if (/\/v1($|\/.+)/.test(trimmed)) return `${trimmed}/chat/completions`;
    return `${trimmed}/v1/chat/completions`;
  },
});

export const callProvider = adapter.callProvider;
export const chatCompletionStream = adapter.chatCompletionStream;
