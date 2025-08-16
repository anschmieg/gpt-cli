/**
 * Permission-safe logging helpers. Reads `GPT_CLI_VERBOSE` if available.
 */
function verboseEnabled(): boolean {
  try {
    return Deno.env.get("GPT_CLI_VERBOSE") === "1";
  } catch {
    return false;
  }
}

export function debug(...args: unknown[]) {
  if (verboseEnabled()) console.log("[DEBUG]", ...args);
}

// NOTE: prefer `debug` explicitly; no `log` alias exported.

export function info(...args: unknown[]) {
  console.log("[INFO]", ...args);
}

export function warn(...args: unknown[]) {
  console.warn("[WARN]", ...args);
}

export function error(...args: unknown[]) {
  console.error("[ERROR]", ...args);
}
