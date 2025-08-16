import {
  assert,
  assertEquals,
} from "https://deno.land/std@0.201.0/testing/asserts.ts";

const adapters = ["openai", "copilot", "gemini"] as const;

for (const name of adapters) {
  Deno.test(`adapter shape: ${name}`, async () => {
    const m = await import(`../../providers/${name}.ts`);
    assert(m, `imported module for provider ${name}`);
    assert(
      typeof m.callProvider === "function",
      `provider ${name} should export callProvider`,
    );
    if (m.chatCompletionStream !== undefined) {
      assertEquals(typeof m.chatCompletionStream, "function");
    }
  });
}
