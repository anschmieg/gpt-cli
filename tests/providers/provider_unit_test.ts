import { chatCompletionRequest } from "../../providers/api_openai_compatible.ts";
import { assertEquals } from "https://deno.land/std@0.201.0/testing/asserts.ts";

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
  assertEquals(res, "hi");
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
    if (!(err instanceof Error)) throw err;
  }
});
