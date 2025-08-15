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
  const callProvider = callProviderFn ??
    ((c: CoreConfig) =>
      import("./providers/openai.ts").then((m) => m.callProvider(c)));
  const render = renderMd ?? renderMarkdown;
  const logFn = logger ?? log;

  if (config.verbose) logFn("Config:", config);
  let response;
  try {
    response = await callProvider(config);
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
  if (config.verbose) logFn("Raw response:", response);
  // Output as markdown or plaintext
  if (response.markdown) {
    console.log(render(response.markdown));
  } else {
    console.log(response.text);
  }
}
