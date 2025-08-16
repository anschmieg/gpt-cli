import {
  chatCompletionRequest,
  chatCompletionRequestStream,
} from "../src/providers/openai_request.ts";
import { throwNormalized } from "../src/utils/adapter_utils.ts";
import type {
  ChatRequest,
  Fetcher,
  ProviderConfig,
  ProviderOptions,
} from "../src/providers/types.ts";

type Descriptor = {
  apiKeyEnv?: string;
  baseEnv?: string;
  defaultBase?: string;
  // when true, adapter will not require an API key (useful for test/no-auth)
  allowNoApiKey?: boolean;
  // optional custom url builder: given resolved base -> final URL
  constructUrl?: (base: string) => string;
};

export function makeOpenAICompatibleAdapter(descriptor: Descriptor) {
  const { apiKeyEnv, baseEnv, defaultBase, constructUrl } = descriptor;

  function buildUrl(base: string) {
    if (constructUrl) return constructUrl(base);
    return `${base.replace(/\/$/, "")}/v1/chat/completions`;
  }

  async function callProvider(
    config: ProviderConfig,
    opts?: ProviderOptions,
  ): Promise<{ text?: string; markdown?: string }> {
    const apiKey = opts?.apiKey ??
      (apiKeyEnv ? Deno.env.get(apiKeyEnv) ?? "" : "");
    if (!apiKey && !descriptor.allowNoApiKey) {
      throwNormalized(
        new Error(`${apiKeyEnv ?? "API_KEY"} not provided in ProviderOptions`),
      );
    }

    const base = opts?.baseUrl ??
      (baseEnv ? Deno.env.get(baseEnv) : undefined) ?? defaultBase ?? "";
    if (!base) {
      throwNormalized(
        new Error(`${baseEnv ?? "BASE_URL"} not provided in ProviderOptions`),
      );
    }

    const url = buildUrl(base);

    const body: ChatRequest = {
      model: config.model,
      messages: (([
        config.system ? { role: "system", content: config.system } : null,
        { role: "user", content: config.prompt ?? "" },
      ].filter(Boolean)) as unknown) as ChatRequest["messages"],
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

  function chatCompletionStream(
    req: ChatRequest,
    baseUrlOrOptions: string | {
      baseUrl?: string;
      fetcher?: Fetcher;
      apiKey?: string;
    } = { baseUrl: defaultBase },
  ): Promise<AsyncGenerator<string, void, unknown>> {
    const opts = typeof baseUrlOrOptions === "string"
      ? { baseUrl: baseUrlOrOptions }
      : baseUrlOrOptions || {};
    const baseUrl = opts.baseUrl ?? defaultBase ??
      (baseEnv ? Deno.env.get(baseEnv) ?? "" : "");
    const fetcher: Fetcher = opts.fetcher ??
      ((input, init) => fetch(input, init));
    const apiKey = opts.apiKey ??
      (apiKeyEnv ? Deno.env.get(apiKeyEnv) ?? "" : "");

    const url = buildUrl(baseUrl);

    const gen = chatCompletionRequestStream({
      url,
      apiKey,
      body: req,
      fetcher,
    });
    return Promise.resolve(gen);
  }

  return { callProvider, chatCompletionStream };
}

export type { Fetcher };
