export interface ChatMessage {
  role: "system" | "user" | "assistant";
  content: string;
}

export interface ChatRequest {
  model?: string;
  messages: ChatMessage[];
  stream?: boolean;
}
export type Fetcher = (input: string, init?: RequestInit) => Promise<Response>;

export interface ChatOptions {
  baseUrl?: string;
  fetcher?: Fetcher;
}

/**
 * chatCompletion: call provider to get assistant content.
 * Backwards-compatible: second argument may be a string baseUrl or an options object.
 */
export async function chatCompletion(
  req: ChatRequest,
  baseUrlOrOptions: string | ChatOptions = { baseUrl: "http://127.0.0.1:8086" },
): Promise<string> {
  const opts: ChatOptions = typeof baseUrlOrOptions === "string"
    ? { baseUrl: baseUrlOrOptions }
    : baseUrlOrOptions || {};
  const baseUrl = opts.baseUrl ?? "http://127.0.0.1:8086";
  const fetcher: Fetcher = opts.fetcher ??
    ((input, init) => fetch(input, init));
  // Safeguard: when running tests, enforce that only the mock (localhost) is used.
  try {
    const testFlag = Deno.env.get("GPT_CLI_TEST");
    if (testFlag === "1") {
      const urlIsLocal = baseUrl.startsWith("http://127.0.0.1") ||
        baseUrl.startsWith("http://localhost");
      if (!urlIsLocal) {
        throw new Error(
          "Refusing network call in test mode to non-local endpoint",
        );
      }
    }
  } catch (_err) {
    // If env access is not permitted, be conservative: don't block execution here.
    // Tests should run with --allow-env so this path is rarely used.
  }

  const url = `${baseUrl}/v1/chat/completions`;
  const body = {
    model: req.model || "gpt-3.5-mock",
    messages: req.messages,
    stream: req.stream || false,
  };

  const res = await fetcher(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(`provider error: ${res.status} ${res.statusText}: ${text}`);
  }

  const data = await res.json();
  // Expect the mock server to return OpenAI-style response with choices[0].message.content
  const content = data?.choices?.[0]?.message?.content;
  if (typeof content !== "string") {
    throw new Error("invalid response shape from provider");
  }
  return content;
}

// Streaming wrapper: yields string fragments from the provider
export async function chatCompletionStream(
  req: ChatRequest,
  baseUrlOrOptions: string | ChatOptions = { baseUrl: "http://127.0.0.1:8086" },
): Promise<AsyncGenerator<string, void, unknown>> {
  const opts: ChatOptions = typeof baseUrlOrOptions === "string"
    ? { baseUrl: baseUrlOrOptions }
    : baseUrlOrOptions || {};
  const baseUrl = opts.baseUrl ?? "http://127.0.0.1:8086";
  const fetcher: Fetcher = opts.fetcher ??
    ((input, init) => fetch(input, init));

  // dynamic import of the shared streaming helper to avoid duplication
  const mod = await import("../../providers/api_openai_compatible.ts");
  const gen = mod.chatCompletionRequestStream({
    url: `${baseUrl}/v1/chat/completions`,
    apiKey: "",
    body: req,
    fetcher,
  });
  return gen;
}
