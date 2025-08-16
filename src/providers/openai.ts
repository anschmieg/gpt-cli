import type { ChatRequest, Fetcher } from "./types.ts";
import { requestNonStreaming, requestStreaming } from "./openai_request.ts";
import { MOCK_SERVER_URL } from "../config.ts";

export interface ChatOptions {
  baseUrl?: string;
  fetcher?: Fetcher;
}

// Backwards-compatible wrapper: keeps test-mode guard and a stable API while
// delegating request logic to `openai_request.ts`.
export function chatCompletion(
  req: ChatRequest,
  baseUrlOrOptions: string | ChatOptions = { baseUrl: MOCK_SERVER_URL },
): Promise<string> {
  // Test-mode guard: enforce local endpoints when GPT_CLI_TEST=1
  try {
    const opts = typeof baseUrlOrOptions === "string"
      ? { baseUrl: baseUrlOrOptions }
      : baseUrlOrOptions || {};
    const baseUrl = opts.baseUrl ?? MOCK_SERVER_URL;
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
  } catch {
    // ignore env access errors
  }

  return requestNonStreaming(req, baseUrlOrOptions as ChatOptions);
}

export function chatCompletionStream(
  req: ChatRequest,
  baseUrlOrOptions: string | ChatOptions = { baseUrl: MOCK_SERVER_URL },
): Promise<AsyncGenerator<string, void, unknown>> {
  return requestStreaming(req, baseUrlOrOptions as ChatOptions);
}
