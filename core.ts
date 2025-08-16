import { log } from "./src/utils/log.ts";
import { renderMarkdown } from "./src/utils/markdown.ts";
import { DEFAULTS, MOCK_SERVER_URL } from "./src/config.ts";

export interface CoreConfig {
  provider?: string;
  model?: string;
  temperature?: number;
  system?: string;
  file?: string;
  verbose?: boolean;
  autoRetryModel?: boolean;
  prompt?: string;
  useMarkdown?: boolean;
  stream?: boolean;
}

import type { ProviderOptions } from "./src/providers/types.ts";
import {
  normalizeProviderError,
  ProviderError,
  validateAdapterModule,
} from "./src/utils/adapter_utils.ts";

export type CallProviderFn = (
  config: CoreConfig,
  opts?: ProviderOptions,
) => Promise<{ text?: string; markdown?: string }>;

export async function runCore(
  config: CoreConfig,
  callProviderFn?: CallProviderFn,
  renderMd?: (s: string) => string,
  logger?: (...args: unknown[]) => void,
  providerOpts?: ProviderOptions,
): Promise<{ ok: true } | { ok: false; error: string }> {
  // Configure defaults
  const defaultSystem =
    `You are an AI assistant called via CLI. Respond concisely and clearly, focusing only on the user's prompt. Include only very brief explanations unless explicitly asked.`;
  const cfg: CoreConfig = {
    provider: config.provider ?? DEFAULTS.provider,
    model: config.model ?? DEFAULTS.model,
    temperature: config.temperature ?? DEFAULTS.temperature,
    system: config.system ?? defaultSystem,
    file: config.file,
    verbose: config.verbose ?? false,
    autoRetryModel: config.autoRetryModel ?? false,
    prompt: config.prompt,
    useMarkdown: config.useMarkdown ?? true,
  };

  const callProvider = callProviderFn ??
    ((c: CoreConfig, opts?: ProviderOptions) => {
      const providerName = (c.provider ?? "openai").toLowerCase();
      // Dynamic import of the provider adapter. The adapter must export
      // `callProvider(config, opts?)`.
      return import(`./adapters/${providerName}.ts`).then((m) => {
        validateAdapterModule(m, providerName);
        return m.callProvider(c, opts);
      });
    });
  const render = renderMd ?? renderMarkdown;
  const logFn = logger ?? log;

  if (cfg.verbose) logFn("Config:", cfg);

  // If streaming was requested, try to consume a provider streaming API first.
  if (cfg.stream) {
    try {
      const providerName = (cfg.provider ?? "openai").toLowerCase();
      const m = await import(`./adapters/${providerName}.ts`);
      try {
        // eslint-disable-next-line no-console
        console.log("runCore: provider module keys:", Object.keys(m));
      } catch {
        // ignore
      }
      if (m && typeof m.chatCompletionStream === "function") {
        // Runtime shape check: ensure basic non-stream call exists too.
        if (typeof m.callProvider !== "function") {
          throw new Error(
            `Provider adapter ./adapters/${providerName}.ts must export a 'callProvider(config, opts?)' function`,
          );
        }
        // Call provider streaming API and print chunks as they arrive.
        const baseUrl = providerOpts?.baseUrl ??
          (Deno.env.get("GPT_CLI_TEST") === "1" ? MOCK_SERVER_URL : undefined);
        const opts = { ...providerOpts, baseUrl } as
          | ProviderOptions
          | undefined;
        const gen: AsyncGenerator<string, void, unknown> = await m
          .chatCompletionStream({
            model: cfg.model,
            messages: [{ role: "user", content: cfg.prompt ?? "" }],
            stream: true,
          }, opts);
        for await (const chunk of gen) {
          // For now, print raw chunks. downstream: pass through render for markdown.
          Deno.stdout.write(new TextEncoder().encode(chunk));
        }
        // finish with newline
        console.log("");
        return { ok: true };
      }
    } catch (err) {
      // If streaming isn't supported or fails, fall back to non-streaming provider below.
      try {
        if (cfg.verbose) {
          if (err instanceof ProviderError) {
            logFn("Streaming provider error code:", err.code);
          }
          logFn("Streaming provider error, falling back:", err);
        }
      } catch {
        // ignore logging errors
      }
    }
  }

  // Non-streaming path: call provider to get full response.
  let response;
  try {
    response = await callProvider(cfg, providerOpts);
  } catch (err) {
    logFn("Provider error:", err);
    const normalized = normalizeProviderError(err);
    let hint = "";
    let modelNotSupported = false;
    try {
      const lower = (normalized.message ?? "").toLowerCase();
      modelNotSupported = normalized.code === "model_not_supported" ||
        lower.includes("model_not_supported") ||
        lower.includes("model is not supported") ||
        lower.includes("requested model is not supported");
      if (modelNotSupported) {
        hint =
          "\nHint: the provider rejected the requested model. Trying again without `--model`...";
      }
    } catch {
      // ignore
    }

    // Automated retry: if provider indicated the model is not supported,
    // and the caller opted in, try once more without sending an explicit model.
    if (modelNotSupported && cfg.autoRetryModel) {
      try {
        const retryCfg = { ...cfg, model: undefined } as CoreConfig;
        if (cfg.verbose) logFn("Retrying provider call without model...");
        response = await callProvider(retryCfg, providerOpts);
      } catch (err2) {
        const normalized2 = normalizeProviderError(err2);
        const out = `${normalized2.message} (after retry)`;
        console.error("Error:", out);
        return { ok: false, error: normalized2.message };
      }
    } else {
      const out = `${normalized.message}${hint}`;
      console.error("Error:", out);
      return { ok: false, error: normalized.message };
    }
  }
  if (cfg.verbose) logFn("Raw response:", response);

  // Respect the useMarkdown flag: if disabled, prefer plain text output.
  if (cfg.useMarkdown === false) {
    // Prefer text, but fall back to markdown content if no text present.
    console.log(response.text ?? response.markdown ?? "");
    return { ok: true };
  }

  // When markdown is enabled, return markdown content without any top-level
  // wrapper. Fall back to text if markdown isn't provided by the provider.
  if (response.markdown) {
    // render may be a noop; we still pass through to allow optional transforms
    // but do not add any wrapper.
    console.log(render(response.markdown));
  } else {
    console.log(response.text ?? "");
  }
  return { ok: true };
}
