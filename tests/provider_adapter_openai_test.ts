import { expect } from "jsr:@std/expect";
import { callProvider as openai } from "../providers/openai.ts";
import type { Fetcher } from "../providers/openai.ts";

Deno.test("openai adapter errors when OPENAI_API_KEY missing", async () => {
  try {
    await openai({ model: "x", prompt: "hi" });
    throw new Error("expected error when OPENAI_API_KEY missing");
  } catch (err) {
    if (!(err instanceof Error)) throw err;
  }
});

Deno.test("openai adapter calls fetcher with correct URL and returns text", async () => {
  let calledUrl = "";
  let calledInit: RequestInit | undefined;
  const fetcher: Fetcher = (input: string, init?: RequestInit) => {
    calledUrl = input;
    calledInit = init;
    const body = { choices: [{ message: { content: "hello from openai" } }] };
    return Promise.resolve(
      new Response(JSON.stringify(body), { status: 200 }),
    );
  };
  const res = await openai({ model: "m", prompt: "p" }, {
    apiKey: "test-key",
    baseUrl: "https://api.example",
    fetcher,
  });
  expect(res.text).toBe("hello from openai");
  expect(calledUrl).toBe("https://api.example/v1/chat/completions");
  // basic header check
  const headers = calledInit?.headers as Record<string, string> | undefined;
  const auth = headers && headers["Authorization"];
  expect(auth).toBe("Bearer test-key");
});
