import { parse } from "https://deno.land/std@0.203.0/flags/mod.ts";
import { runCore } from "./core.ts";
import { DEFAULTS } from "./src/config.ts";
import type { ProviderOptions } from "./src/providers/types.ts";

function printHelp() {
  console.log(
    `gpt-cli: Portable GPT API Wrapper\n\nUsage: gpt-cli [options] <prompt>\n\nOptions:\n  --provider   API provider (openai, gemini, etc)\n  --model      Model name\n  --temperature Temperature (float)\n  --system     System prompt\n  --file       File to upload\n  --verbose    Enable verbose logging\n  -h, --help   Show help\n`,
  );
}

export function parseArgs(argv: string[]) {
  return parse(argv, {
    string: ["provider", "model", "system", "file"],
    boolean: ["verbose", "help", "markdown", "retry-model"],
    default: {
      provider: DEFAULTS.provider,
      model: undefined,
      temperature: String(DEFAULTS.temperature),
      verbose: DEFAULTS.verbose,
      markdown: DEFAULTS.markdown,
      "retry-model": false,
    },
    alias: { h: "help" },
  });
}

if (import.meta.main) {
  const args = parseArgs(Deno.args);

  if (args.help || args._.length === 0) {
    printHelp();
    Deno.exit(0);
  }

  const config = {
    provider: args.provider,
    model: args.model,
    temperature: parseFloat(args.temperature as string),
    system: args.system,
    file: args.file,
    verbose: args.verbose,
    useMarkdown: Boolean(args.markdown),
    autoRetryModel: Boolean(args["retry-model"]),
    prompt: args._.join(" "),
  };

  // Build provider options from environment at the CLI boundary so library code
  // doesn't need to read environment variables directly. This keeps core and
  // providers testable without Deno env permissions.
  const providerName = (config.provider ?? "openai").toLowerCase();
  let providerOpts: ProviderOptions | undefined = undefined;
  if (providerName === "openai") {
    providerOpts = {
      apiKey: Deno.env.get("OPENAI_API_KEY") ?? undefined,
      baseUrl: Deno.env.get("OPENAI_API_BASE") ?? undefined,
    };
  } else if (providerName === "copilot") {
    providerOpts = {
      apiKey: Deno.env.get("COPILOT_API_KEY") ?? undefined,
      baseUrl: Deno.env.get("COPILOT_API_BASE") ?? undefined,
    };
  } else if (providerName === "gemini") {
    providerOpts = {
      apiKey: Deno.env.get("GEMINI_API_KEY") ?? undefined,
      baseUrl: undefined,
    };
  }

  const res = await runCore(
    config,
    undefined,
    undefined,
    undefined,
    providerOpts,
  );
  if (!res || (res as { ok: boolean }).ok === false) {
    const err = (res as { ok: false; error: string } | undefined)?.error ??
      "unknown error";
    console.error("Error:", err);
    Deno.exit(1);
  }
}
