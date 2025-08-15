import { chatCompletion } from "../src/providers/openai.ts";

Deno.test("chatCompletion returns mocked content without network", async () => {
  const fakeFetch = (_url: string, _init?: RequestInit) => {
    return Promise.resolve(
      new Response(
        JSON.stringify({
          choices: [{
            message: { content: "pure mock result", role: "assistant" },
          }],
        }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      ),
    );
  };

  const out = await chatCompletion({
    messages: [{ role: "user", content: "hi" }],
  }, { fetcher: fakeFetch });
  if (out !== "pure mock result") throw new Error(`unexpected: ${out}`);
});

Deno.test("chatCompletion surfaces provider errors", async () => {
  const badFetch = (_url: string, _init?: RequestInit) => {
    return Promise.resolve(new Response("bad", { status: 500 }));
  };

  let threw = false;
  try {
    await chatCompletion({ messages: [{ role: "user", content: "hi" }] }, {
      fetcher: badFetch,
    });
  } catch (e) {
    threw = true;
    if (!/provider error/.test(String(e))) throw e;
  }
  if (!threw) throw new Error("expected error but none thrown");
});
