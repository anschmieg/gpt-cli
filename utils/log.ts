export function log(...args: unknown[]) {
  if (Deno.env.get("GPT_CLI_VERBOSE") === "1") {
    console.log("[DEBUG]", ...args);
  }
}
