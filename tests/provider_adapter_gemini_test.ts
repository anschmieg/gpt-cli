import { assertEquals } from "jsr:@std/assert/equals";
import { callProvider as gemini } from "../providers/gemini.ts";
import type { Fetcher } from "../providers/gemini.ts";

Deno.test("gemini adapter errors when GEMINI_API_KEY missing", async () => {
  try {
    await gemini({ model: "x", prompt: "hi" });
    throw new Error("expected error when GEMINI_API_KEY missing");
  } catch (err) {
    if (!(err instanceof Error)) throw err;
  }
});

Deno.test("gemini adapter calls fetcher with correct URL and returns text", async () => {
  let calledUrl = "";
  let calledInit: RequestInit | undefined;
  const fetcher: Fetcher = (input: string, init?: RequestInit) => {
    calledUrl = input;
    calledInit = init;
    const body = { choices: [{ message: { content: "hello from gemini" } }] };
    return Promise.resolve(
      new Response(JSON.stringify(body), { status: 200 }),
    );
  };
  const res = await gemini({ model: "m", prompt: "p" }, {
    apiKey: "g-key",
    fetcher,
  });
  assertEquals(res.text, "hello from gemini");
  assertEquals(
    calledUrl,
    "https://generativelanguage.googleapis.com/v1beta/openai",
  );
  const headers = calledInit?.headers as Record<string, string> | undefined;
  const auth = headers && headers["Authorization"];
  assertEquals(auth, "Bearer g-key");
});
