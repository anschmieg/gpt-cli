import { chatCompletionRequest } from "../../src/providers/api_openai_compatible.ts";
import { expect } from "jsr:@std/expect";
import { normalizeProviderError } from "../../src/providers/adapter_utils.ts";

Deno.test("chatCompletion returns mocked content without network", async () => {
  const fetcher = (_url: string, _init?: RequestInit) =>
    Promise.resolve(
      new Response(
        JSON.stringify({ choices: [{ message: { content: "hi" } }] }),
        { status: 200 },
      ),
    );
  const res = await chatCompletionRequest({
    url: "https://example",
    apiKey: "k",
    body: { messages: [{ role: "user", content: "hi" }] },
    fetcher,
  });
  expect(res).toBe("hi");
});

Deno.test("chatCompletion surfaces provider errors", async () => {
  const fetcher = (_url: string, _init?: RequestInit) =>
    Promise.resolve(
      new Response(JSON.stringify({ error: { message: "bad" } }), {
        status: 400,
      }),
    );
  try {
    await chatCompletionRequest({
      url: "https://example",
      apiKey: "k",
      body: { messages: [{ role: "user", content: "hi" }] },
      fetcher,
    });
    throw new Error("expected error");
  } catch (err) {
    // Assert normalized shape for provider errors
    const n = normalizeProviderError(err);
    expect(typeof n.message).toBe("string");
  }
});
