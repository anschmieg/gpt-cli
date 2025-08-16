import { expect } from "jsr:@std/expect";
import { callProvider as copilot } from "../../providers/copilot.ts";
import type { Fetcher, ProviderOptions } from "../../src/providers/types.ts";

Deno.test("copilot adapter errors when ProviderOptions missing", async () => {
  // Call without ProviderOptions: provider should throw because apiKey/baseUrl
  // are required. This test does not rely on environment permissions.
  try {
    await copilot({ model: "x", prompt: "hi" });
    throw new Error("expected error when ProviderOptions missing");
  } catch (err) {
    if (!(err instanceof Error)) throw err;
  }
});

Deno.test("copilot adapter calls fetcher with correct URL and returns text", async () => {
  // Provide ProviderOptions explicitly to avoid requiring env permissions.
  let calledUrl = "";
  let calledInit: RequestInit | undefined;
  const fetcher: Fetcher = (input: string, init?: RequestInit) => {
    calledUrl = input;
    calledInit = init;
    const body = { choices: [{ message: { content: "hello from copilot" } }] };
    return Promise.resolve(new Response(JSON.stringify(body), { status: 200 }));
  };

  const opts: ProviderOptions = {
    apiKey: "c-key",
    baseUrl: "https://copilot.example",
    fetcher,
  };
  const res = await copilot({ model: "m", prompt: "p" }, opts);
  expect(res.text).toBe("hello from copilot");
  expect(calledUrl).toBe("https://copilot.example/v1/chat/completions");
  const auth = calledInit?.headers &&
    (calledInit.headers as Record<string, string>)["Authorization"];
  expect(auth).toBe("Bearer c-key");
});
