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
) {
  const apiKey = opts?.apiKey ?? "";
  if (!apiKey) {
    throwNormalized(
      new Error("COPILOT_API_KEY not provided in ProviderOptions"),
    );
  }
  const base = opts?.baseUrl ?? Deno.env.get("COPILOT_API_BASE") ?? "";
  if (!base) {
    throwNormalized(
      new Error("COPILOT_API_BASE not provided in ProviderOptions"),
    );
  }
  // Normalize URL construction:
  // - If base already ends with '/v1/chat/completions', use it as-is.
  // - If base contains '/v1' but not the full path, append '/chat/completions'.
  // - Otherwise append '/v1/chat/completions'. This avoids duplicating '/v1'.
  let url: string;
  try {
    const trimmed = base.replace(/\/$/, "");
    if (/\/v1\/chat\/completions$/.test(trimmed)) {
      url = trimmed;
    } else if (/\/v1($|\/.+)/.test(trimmed)) {
      url = `${trimmed}/chat/completions`;
    } else {
      url = `${trimmed}/v1/chat/completions`;
    }
  } catch {
    // Fallback: naive append
    url = `${base.replace(/\/$/, "")}/v1/chat/completions`;
  }

  const body: ChatRequest = {
    model: config.model,
    messages: ([
      config.system ? { role: "system", content: config.system } : null,
      { role: "user", content: config.prompt ?? "" },
    ].filter(Boolean)) as ChatRequest["messages"],
    stream: false,
  };

  // Optional debug: set GPT_CLI_DEBUG=1 to print request details when troubleshooting.
  try {
    if (Deno.env.get("GPT_CLI_DEBUG") === "1") {
      // eslint-disable-next-line no-console
      console.log("copilot.callProvider: url=", url, "body=", body);
    }
  } catch {
    // ignore logging permission errors
  }

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
