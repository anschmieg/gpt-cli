/**
 * runCli - minimal CLI core used by main and tests.
 * Returns a greeting string for now; later it will build requests and call providers.
 */
export function runCli(args: string[] = []): Promise<string> {
  // Keep logic simple for tests. Real parsing lives in `cli.ts`/`core.ts` later.
  const greeting = "Hello from Deno CLI!";
  // If a prompt argument was passed, echo it for basic behavior.
  if (args.length > 0) {
    return Promise.resolve(`${greeting} Args: ${args.join(" ")}`);
  }
  return Promise.resolve(greeting);
}
