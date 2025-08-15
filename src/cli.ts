/**
 * runCli - minimal CLI core used by main and tests.
 * Returns a greeting string for now; later it will build requests and call providers.
 */
import { parse as parseFlags } from "https://deno.land/std@0.201.0/flags/mod.ts";

export function parseArgs(argv: string[] = []) {
  // Lightweight wrapper around std/flags for predictable parsing in tests.
  const parsed = parseFlags(argv, {
    boolean: ["help", "verbose"],
    string: ["format", "count"],
    alias: { h: "help", v: "verbose", f: "format", c: "count" },
    default: { verbose: false },
  });

  const { _, help, verbose, format, count } = parsed as unknown as ReturnType<
    typeof parseFlags
  >;
  const countNum = count !== undefined ? Number(count) : undefined;

  return {
    help: Boolean(help),
    verbose: Boolean(verbose),
    format: format ?? undefined,
    count: Number.isFinite(countNum) ? countNum : undefined,
    positional: Array.isArray(_) ? _.map(String) : [],
  } as const;
}

export function runCli(args: string[] = []): Promise<string> {
  // Keep existing behavior for backwards compatibility with tests.
  const greeting = "Hello from Deno CLI!";
  if (args.length > 0) {
    return Promise.resolve(`${greeting} Args: ${args.join(" ")}`);
  }
  return Promise.resolve(greeting);
}
