import { assertEquals } from "https://deno.land/std@0.201.0/testing/asserts.ts";
import { callProvider as gemini } from "../../providers/gemini.ts";
import type { Fetcher, ProviderOptions } from "../../src/providers/types.ts";

Deno.test("gemini adapter errors when ProviderOptions missing", async () => {
  try {
    await gemini({ model: "x", prompt: "hi" });
    throw new Error("expected error when ProviderOptions missing");
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
    return Promise.resolve(new Response(JSON.stringify(body), { status: 200 }));
  };

  const opts: ProviderOptions = { apiKey: "g-key", fetcher };
  const res = await gemini({ model: "m", prompt: "p" }, opts);
  assertEquals(res.text, "hello from gemini");
  assertEquals(
    calledUrl,
    "https://generativelanguage.googleapis.com/v1beta/openai",
  );
  const auth = calledInit?.headers &&
    (calledInit.headers as Record<string, string>)["Authorization"];
  assertEquals(auth, "Bearer g-key");
});
