import { expect } from "jsr:@std/expect";
import { normalizeProviderError } from "../../src/utils/adapter_utils.ts";

Deno.test("normalizeProviderError handles string", () => {
  const n = normalizeProviderError("oh no");
  expect(n.message).toBe("oh no");
});

Deno.test("normalizeProviderError handles Error", () => {
  const n = normalizeProviderError(new Error("boom"));
  expect(n.message).toBe("boom");
});

Deno.test("normalizeProviderError handles nested error object", () => {
  const n = normalizeProviderError({ error: { message: "bad", code: "x" } });
  expect(n.message).toBe("bad");
  expect(n.code).toBe("x");
});

Deno.test("normalizeProviderError handles status/statusText", () => {
  const n = normalizeProviderError({ status: 404, statusText: "Not Found" });
  expect(n.message.includes("HTTP 404")).toBe(true);
});
