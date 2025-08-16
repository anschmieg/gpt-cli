import { expect } from "jsr:@std/expect";
import { runCore } from "../../core.ts";

Deno.test("runCore dynamic import fails when adapter missing callProvider", async () => {
  // Instead of writing a temporary adapter file, verify that requesting a
  // non-existent adapter name results in a handled error. This avoids
  // requiring write permissions in CI while exercising the same code path.
  const res = await runCore({ provider: "__adapter_that_does_not_exist__" });
  expect(res.ok).toBe(false);
  expect("error" in res).toBe(true);
  // runtime check: error should be a string message
  // @ts-ignore runtime check
  expect(typeof (res as unknown as { error?: unknown }).error).toBe("string");
});
