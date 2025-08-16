import { chatCompletion } from "../../src/providers/openai.ts";
import { MOCK_SERVER_URL } from "../../src/config.ts";

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
  }, { baseUrl: MOCK_SERVER_URL, fetcher });
  if (res !== "hello") throw new Error("expected hello");
});
