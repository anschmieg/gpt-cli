import type { ChatRequest, Fetcher } from "./types.ts";
import { ensureResponseOk } from "../utils/adapter_utils.ts";
import { parseSSEStream } from "../utils/sse_parser.ts";

type Opts = { baseUrl?: string; fetcher?: Fetcher; apiKey?: string };

// --- Canonical OpenAI-compatible low-level helpers (merged from
// src/providers/api_openai_compatible.ts) ---
export async function chatCompletionRequest(
  params: {
    url: string;
    apiKey: string;
    body: ChatRequest;
    fetcher?: Fetcher;
  },
): Promise<string> {
  const { url, apiKey, body, fetcher } = params;
  const res = await postAndEnsure(url, apiKey, body, fetcher);

  const data = await res.json();
  const content = data?.choices?.[0]?.message?.content;
  if (typeof content !== "string") {
    throw new Error("invalid response shape from provider");
  }
  return content;
}

// Streaming variant: returns an async generator that yields string chunks as they arrive
export async function* chatCompletionRequestStream(
  params: {
    url: string;
    apiKey: string;
    body: ChatRequest;
    fetcher?: Fetcher;
  },
): AsyncGenerator<string, void, unknown> {
  const { url, apiKey, body, fetcher } = params;
  const res = await postAndEnsure(url, apiKey, body, fetcher);

  const reader = res.body?.getReader();
  if (!reader) return;
  try {
    for await (const chunk of parseSSEStream(reader)) {
      yield chunk;
    }
  } finally {
    try {
      await reader.cancel();
    } catch {
      // ignore
    }
  }
}

async function postAndEnsure(
  url: string,
  apiKey: string,
  body: ChatRequest,
  fetcher?: Fetcher,
): Promise<Response> {
  const realFetch: Fetcher = fetcher ?? ((input, init) => fetch(input, init));
  const res = await realFetch(url, {
    method: "POST",
    headers: {
      "Authorization": `Bearer ${apiKey}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });
  await ensureResponseOk(res);
  return res;
}

// --- Higher-level helpers (previously in this file) ---
export async function requestNonStreaming(
  req: ChatRequest,
  baseUrlOrOptions: string | Opts = { baseUrl: "http://127.0.0.1:8086" },
): Promise<string> {
  const opts: Opts = typeof baseUrlOrOptions === "string"
    ? { baseUrl: baseUrlOrOptions }
    : baseUrlOrOptions || {};
  const baseUrl = opts.baseUrl ?? "http://127.0.0.1:8086";
  const fetcher: Fetcher = opts.fetcher ??
    ((input, init) => fetch(input, init));
  const apiKey = opts.apiKey ?? "";

  const url = `${baseUrl.replace(/\/$/, "")}/v1/chat/completions`;
  return await chatCompletionRequest({ url, apiKey, body: req, fetcher });
}

export function requestStreaming(
  req: ChatRequest,
  baseUrlOrOptions: string | Opts = { baseUrl: "http://127.0.0.1:8086" },
): Promise<AsyncGenerator<string, void, unknown>> {
  const opts: Opts = typeof baseUrlOrOptions === "string"
    ? { baseUrl: baseUrlOrOptions }
    : baseUrlOrOptions || {};
  const baseUrl = opts.baseUrl ?? "http://127.0.0.1:8086";
  const fetcher: Fetcher = opts.fetcher ??
    ((input, init) => fetch(input, init));
  const apiKey = opts.apiKey ?? "";

  const url = `${baseUrl.replace(/\/$/, "")}/v1/chat/completions`;
  return Promise.resolve(
    chatCompletionRequestStream({ url, apiKey, body: req, fetcher }),
  );
}
