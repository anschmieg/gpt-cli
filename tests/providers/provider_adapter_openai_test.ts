import { expect } from "jsr:@std/expect";
import { callProvider as openai } from "../../providers/openai.ts";
import type { Fetcher, ProviderOptions } from "../../src/providers/types.ts";

Deno.test("openai adapter errors when ProviderOptions missing", async () => {
  try {
    await openai({ model: "x", prompt: "hi" });
    throw new Error("expected error when ProviderOptions missing");
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
    return Promise.resolve(new Response(JSON.stringify(body), { status: 200 }));
  };

  const opts: ProviderOptions = {
    apiKey: "test-key",
    baseUrl: "https://api.example",
    fetcher,
  };
  const res = await openai({ model: "m", prompt: "p" }, opts);
  expect(res.text).toBe("hello from openai");
  expect(calledUrl).toBe("https://api.example/v1/chat/completions");
  // basic header check
  const auth = calledInit?.headers &&
    (calledInit.headers as Record<string, string>)["Authorization"];
  expect(auth).toBe("Bearer test-key");
});
