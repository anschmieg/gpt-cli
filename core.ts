import { log } from "./utils/log.ts";
import { renderMarkdown } from "./utils/markdown.ts";

export interface CoreConfig {
  provider?: string;
  model?: string;
  temperature?: number;
  system?: string;
  file?: string;
  verbose?: boolean;
  prompt?: string;
  useMarkdown?: boolean;
}

export type CallProviderFn = (
  config: CoreConfig,
) => Promise<{ text?: string; markdown?: string }>;

export async function runCore(
  config: CoreConfig,
  callProviderFn?: CallProviderFn,
  renderMd?: (s: string) => string,
  logger?: (...args: unknown[]) => void,
) {
  // Configure defaults
  const defaultSystem =
    `You are an AI assistant called via CLI. Respond concisely and clearly, focusing only on the user's prompt. Include only very brief explanations unless explicitly asked.`;
  const cfg: CoreConfig = {
    provider: config.provider ?? "copilot",
    model: config.model ?? "gpt-4.1-mini",
    temperature: config.temperature ?? 0.6,
    system: config.system ?? defaultSystem,
    file: config.file,
    verbose: config.verbose ?? false,
    prompt: config.prompt,
    useMarkdown: config.useMarkdown ?? true,
  };

  const callProvider = callProviderFn ?? ((c: CoreConfig) => {
    const providerName = (c.provider ?? "openai").toLowerCase();
    // Dynamic import of the provider adapter. The adapter must export
    // `callProvider(config, fetcher?)`.
    return import(`./providers/${providerName}.ts`).then((m) =>
      m.callProvider(c)
    );
  });
  const render = renderMd ?? renderMarkdown;
  const logFn = logger ?? log;

  if (cfg.verbose) logFn("Config:", cfg);
  let response;
  try {
    response = await callProvider(cfg);
  } catch (err) {
    logFn("Provider error:", err);
    let msg = String(err);
    if (
      err && typeof err === "object" && (err as Record<string, unknown>).message
    ) {
      msg = String((err as Record<string, unknown>).message);
    }
    console.error("Error:", msg);
    Deno.exit(1);
  }
  if (cfg.verbose) logFn("Raw response:", response);

  // Respect the useMarkdown flag: if disabled, prefer plain text output.
  if (cfg.useMarkdown === false) {
    // Prefer text, but fall back to markdown content if no text present.
    console.log(response.text ?? response.markdown ?? "");
    return;
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
}
