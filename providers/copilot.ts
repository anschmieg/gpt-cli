import { chatCompletionRequest } from "./api_openai_compatible.ts";
import type { ChatRequest } from "./api_openai_compatible.ts";

export type ProviderConfig = {
  model?: string;
  system?: string;
  prompt?: string;
  temperature?: number;
};
export type Fetcher = (input: string, init?: RequestInit) => Promise<Response>;

export async function callProvider(config: ProviderConfig, fetcher?: Fetcher) {
  const apiKey = Deno.env.get("COPILOT_API_KEY");
  if (!apiKey) throw new Error("COPILOT_API_KEY not set in environment");
  const base = Deno.env.get("COPILOT_API_BASE");
  if (!base) throw new Error("COPILOT_API_BASE not set in environment");
  const url = `${base}/v1/chat/completions`;

  const body: ChatRequest = {
    model: config.model,
    messages: ([
      config.system ? { role: "system", content: config.system } : null,
      { role: "user", content: config.prompt ?? "" },
    ].filter(Boolean)) as ChatRequest["messages"],
    stream: false,
  };

  const text = await chatCompletionRequest({ url, apiKey, body, fetcher });
  return { text };
}
