import type { ProviderAdapter } from "../providers/types.ts";

export function validateAdapterModule(m: unknown, providerName = "<unknown>") {
  const mod = m as Partial<ProviderAdapter> | undefined;
  if (!mod || typeof mod.callProvider !== "function") {
    throw new Error(
      `Provider adapter ./adapters/${providerName}.ts must export a 'callProvider(config, opts?)' function`,
    );
  }
  if (
    mod.chatCompletionStream !== undefined &&
    typeof mod.chatCompletionStream !== "function"
  ) {
    throw new Error(
      `Provider adapter ./adapters/${providerName}.ts exported 'chatCompletionStream' but it is not a function`,
    );
  }
  return true;
}

export function normalizeProviderError(
  err: unknown,
): { code?: string; message: string } {
  if (err === null || err === undefined) return { message: "unknown error" };
  if (typeof err === "string") return { message: err };
  if (err instanceof Error) {
    // If Error has a code property (some providers include it), preserve it.
    const anyErr = err as Error & { code?: string };
    return { code: anyErr.code, message: anyErr.message };
  }

  // Try to handle common provider error shapes
  try {
    const obj = err as Record<string, unknown>;
    // shape: { error: { message, code } }
    if (obj.error && typeof obj.error === "object") {
      const e = obj.error as Record<string, unknown>;
      const message =
        (e.message ?? e.msg ?? e.detail ?? JSON.stringify(e)) as string;
      const code = (e.code ?? e.type) as string | undefined;
      return { code, message };
    }
    // shape: { code, message }
    if (obj.message || obj.code) {
      return {
        code: obj.code as string | undefined,
        message: String(obj.message ?? JSON.stringify(obj)),
      };
    }
    // Response-like: { status, statusText }
    if (("status" in obj) && ("statusText" in obj)) {
      const msg = `HTTP ${obj["status"] ?? "?"} ${obj["statusText"] ?? ""}`;
      return { message: msg };
    }
    // Fallback to JSON string
    return { message: JSON.stringify(obj) };
  } catch {
    return { message: String(err) };
  }
}

// Small helper to centralize response -> throw behavior. When a fetch
// Response is not ok, read the body and throw the parsed value (or a
// small object) so callers can pass the thrown value to
// `normalizeProviderError` for consistent messaging.
export async function ensureResponseOk(res: Response): Promise<void> {
  if (res.ok) return;
  const text = await res.text();
  try {
    const parsed = JSON.parse(text);
    throwNormalized(parsed);
  } catch {
    throwNormalized({
      status: res.status,
      statusText: res.statusText,
      body: text,
    });
  }
}

// ProviderError: always thrown by adapters when normalizing errors. Keeps
// the original error available while providing a stable `message` and
// optional `code` property for consumers.
export class ProviderError extends Error {
  code?: string;
  original?: unknown;
  constructor(message: string, code?: string, original?: unknown) {
    super(message);
    this.code = code;
    this.original = original;
    // Maintain proper prototype chain for instanceof checks
    Object.setPrototypeOf(this, ProviderError.prototype);
  }
}

// Throw a ProviderError created from an arbitrary thrown value.
export function throwNormalized(err: unknown): never {
  const n = normalizeProviderError(err);
  throw new ProviderError(n.message, n.code, err);
}
