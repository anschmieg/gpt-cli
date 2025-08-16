import type { ChatRequest, Fetcher } from "./types.ts";
import { MOCK_SERVER_URL } from "../config.ts";
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
  try {
    // Helpful debug output for troubleshooting runtime URL/body issues.
    // Use the GPT_CLI_VERBOSE env var to opt-in so we don't leak secrets by default.
    if (Deno.env.get("GPT_CLI_VERBOSE") === "1") {
      try {
        // Log only the URL and model (do not print apiKey or full body).
        // model may be in body.model or body?.model in some shapes.
        // eslint-disable-next-line no-console
        const bodyUnknown = body as unknown as Record<string, unknown> | null;
        const maybeModel = bodyUnknown ? bodyUnknown["model"] : undefined;
        console.log("[DEBUG] POST ->", url, "model:", String(maybeModel));
      } catch {
        // ignore logging errors
      }
    }
  } catch {
    // ignore environment access errors
  }
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
  baseUrlOrOptions: string | Opts = { baseUrl: MOCK_SERVER_URL },
): Promise<string> {
  const opts: Opts = typeof baseUrlOrOptions === "string"
    ? { baseUrl: baseUrlOrOptions }
    : baseUrlOrOptions || {};
  const baseUrl = opts.baseUrl ?? MOCK_SERVER_URL;
  const fetcher: Fetcher = opts.fetcher ??
    ((input, init) => fetch(input, init));
  const apiKey = opts.apiKey ?? "";

  const url = `${baseUrl.replace(/\/$/, "")}/v1/chat/completions`;
  return await chatCompletionRequest({ url, apiKey, body: req, fetcher });
}

export function requestStreaming(
  req: ChatRequest,
  baseUrlOrOptions: string | Opts = { baseUrl: MOCK_SERVER_URL },
): Promise<AsyncGenerator<string, void, unknown>> {
  const opts: Opts = typeof baseUrlOrOptions === "string"
    ? { baseUrl: baseUrlOrOptions }
    : baseUrlOrOptions || {};
  const baseUrl = opts.baseUrl ?? MOCK_SERVER_URL;
  const fetcher: Fetcher = opts.fetcher ??
    ((input, init) => fetch(input, init));
  const apiKey = opts.apiKey ?? "";

  const url = `${baseUrl.replace(/\/$/, "")}/v1/chat/completions`;
  return Promise.resolve(
    chatCompletionRequestStream({ url, apiKey, body: req, fetcher }),
  );
}
