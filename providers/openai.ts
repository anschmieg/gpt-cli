import { chatCompletionRequest } from "./api_openai_compatible.ts";
import { throwNormalized } from "../src/providers/adapter_utils.ts";
import type {
  ChatRequest,
  Fetcher,
  ProviderConfig,
  ProviderOptions,
} from "../src/providers/types.ts";

export async function callProvider(
  config: ProviderConfig,
  opts?: ProviderOptions,
): Promise<{ text?: string; markdown?: string }> {
  const apiKey = opts?.apiKey ?? "";
  if (!apiKey) {
    throwNormalized(
      new Error("OPENAI_API_KEY not provided in ProviderOptions"),
    );
  }
  const base = opts?.baseUrl ?? Deno.env.get("OPENAI_API_BASE") ??
    "https://api.openai.com";
  const url = `${base.replace(/\/$/, "")}/v1/chat/completions`;

  const body: ChatRequest = {
    model: config.model,
    messages: ([
      config.system ? { role: "system", content: config.system } : null,
      { role: "user", content: config.prompt ?? "" },
    ].filter(Boolean)) as ChatRequest["messages"],
    stream: false,
  };

  const usedFetcher: Fetcher = opts?.fetcher ??
    ((input, init) => fetch(input, init));
  const text = await chatCompletionRequest({
    url,
    apiKey,
    body,
    fetcher: usedFetcher,
  });
  return { text };
}

// Streaming wrapper for provider adapters. Returns an AsyncGenerator<string>
// that yields chunks from the provider.
export async function chatCompletionStream(
  req: ChatRequest,
  baseUrlOrOptions: string | {
    baseUrl?: string;
    fetcher?: Fetcher;
    apiKey?: string;
  } = {
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
  const apiKey = opts.apiKey ?? "";

  const gen = mod.chatCompletionRequestStream({
    url: `${baseUrl.replace(/\/$/, "")}/v1/chat/completions`,
    apiKey,
    body: req,
    fetcher,
  });
  return gen;
}

export type { Fetcher };
