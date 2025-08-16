import { chatCompletionRequest } from "./api_openai_compatible.ts";
import type { ChatRequest } from "./api_openai_compatible.ts";

export type ProviderConfig = {
  model?: string;
  system?: string;
  prompt?: string;
  temperature?: number;
};

export type Fetcher = (input: string, init?: RequestInit) => Promise<Response>;

export async function callProvider(
  config: ProviderConfig,
  fetcher?: Fetcher,
): Promise<{ text?: string; markdown?: string }> {
  const apiKey = Deno.env.get("OPENAI_API_KEY");
  if (!apiKey) throw new Error("OPENAI_API_KEY not set in environment");
  const base = Deno.env.get("OPENAI_API_BASE") ?? "https://api.openai.com";
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

// Streaming wrapper for provider adapters. Returns an AsyncGenerator<string>
// that yields chunks from the provider.
export async function chatCompletionStream(
  req: ChatRequest,
  baseUrlOrOptions: string | { baseUrl?: string; fetcher?: Fetcher } = {
    baseUrl: "http://127.0.0.1:8086",
  },
): Promise<AsyncGenerator<string, void, unknown>> {
  const opts = typeof baseUrlOrOptions === "string"
    ? { baseUrl: baseUrlOrOptions }
    : baseUrlOrOptions || {};
  const baseUrl = opts.baseUrl ?? "http://127.0.0.1:8086";
  const fetcher: Fetcher = opts.fetcher ??
    ((input, init) => fetch(input, init));

  const mod = await import("./api_openai_compatible.ts");
  // Debug: log resolved base URL when running in tests
  try {
    // eslint-disable-next-line no-console
    console.log(
      "openai.chatCompletionStream baseUrl=",
      baseUrl,
      "OPENAI_API_KEY set=",
      Boolean(Deno.env.get("OPENAI_API_KEY")),
    );
  } catch {
    // ignore logging errors in restricted tests
  }
  const gen = mod.chatCompletionRequestStream({
    url: `${baseUrl}/v1/chat/completions`,
    apiKey: Deno.env.get("OPENAI_API_KEY") ?? "",
    body: req,
    fetcher,
  });
  return gen;
}
