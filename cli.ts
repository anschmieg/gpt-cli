import { parse } from "https://deno.land/std@0.203.0/flags/mod.ts";
import { runCore } from "./core.ts";

function printHelp() {
  console.log(
    `gpt-cli: Portable GPT API Wrapper\n\nUsage: gpt-cli [options] <prompt>\n\nOptions:\n  --provider   API provider (openai, gemini, etc)\n  --model      Model name\n  --temperature Temperature (float)\n  --system     System prompt\n  --file       File to upload\n  --verbose    Enable verbose logging\n  -h, --help   Show help\n`,
  );
}

export function parseArgs(argv: string[]) {
  return parse(argv, {
    string: ["provider", "model", "system", "file"],
    boolean: ["verbose", "help", "markdown"],
    default: {
      provider: "copilot",
      model: "gpt-4.1-mini",
      temperature: "0.6",
      verbose: false,
      markdown: true,
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
    prompt: args._.join(" "),
  };

  runCore(config);
}
