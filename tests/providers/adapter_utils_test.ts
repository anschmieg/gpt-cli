import { assertEquals } from "jsr:@std/assert/equals";
import { normalizeProviderError } from "../../src/providers/adapter_utils.ts";

Deno.test("normalizeProviderError handles string", () => {
  const n = normalizeProviderError("oh no");
  assertEquals(n.message, "oh no");
});

Deno.test("normalizeProviderError handles Error", () => {
  const n = normalizeProviderError(new Error("boom"));
  assertEquals(n.message, "boom");
});

Deno.test("normalizeProviderError handles nested error object", () => {
  const n = normalizeProviderError({ error: { message: "bad", code: "x" } });
  assertEquals(n.message, "bad");
  assertEquals(n.code, "x");
});

Deno.test("normalizeProviderError handles status/statusText", () => {
  const n = normalizeProviderError({ status: 404, statusText: "Not Found" });
  assertEquals(n.message.includes("HTTP 404"), true);
});
