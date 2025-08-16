export function log(...args: unknown[]) {
  try {
    if (Deno.env.get("GPT_CLI_VERBOSE") === "1") {
      console.log("[DEBUG]", ...args);
    }
  } catch {
    // ignore when env access is not permitted in tests
  }
}
