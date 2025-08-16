import { expect } from "jsr:@std/expect";

const adapters = ["openai", "copilot", "gemini"] as const;

for (const name of adapters) {
  Deno.test(`adapter shape: ${name}`, async () => {
    const m = await import(`../../providers/${name}.ts`);
    expect(m).toBeTruthy();
    expect(typeof m.callProvider).toBe("function");
    if (m.chatCompletionStream !== undefined) {
      expect(typeof m.chatCompletionStream).toBe("function");
    }
  });
}
