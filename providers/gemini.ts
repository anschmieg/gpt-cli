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
      new Error("GEMINI_API_KEY not provided in ProviderOptions"),
    );
  }
  // Hard-coded OpenAI-compatible Gemini endpoint
  const base = opts?.baseUrl ??
    "https://generativelanguage.googleapis.com/v1beta/openai";
  const url = `${base.replace(/\/$/, "")}`;

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
