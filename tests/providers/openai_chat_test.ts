import { chatCompletion } from "../../src/providers/openai.ts";

Deno.test("openai.chatCompletion returns content from fetcher", async () => {
  const fetcher = (_url: string, _init?: RequestInit) =>
    Promise.resolve(
      new Response(
        JSON.stringify({ choices: [{ message: { content: "hello" } }] }),
        { status: 200 },
      ),
    );
  const res = await chatCompletion({
    model: "m",
    messages: [{ role: "user", content: "hi" }],
  }, { baseUrl: "http://127.0.0.1:8086", fetcher });
  if (res !== "hello") throw new Error("expected hello");
});
