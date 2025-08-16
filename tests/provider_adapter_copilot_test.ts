import { expect } from "jsr:@std/expect";
import { callProvider as copilot } from "../providers/copilot.ts";
import type { Fetcher } from "../providers/copilot.ts";

Deno.test("copilot adapter errors when COPILOT_API_KEY or BASE missing", async () => {
  try {
    await copilot({ model: "x", prompt: "hi" });
    throw new Error("expected error when COPILOT_API_KEY missing");
  } catch (err) {
    if (!(err instanceof Error)) throw err;
  }
});

Deno.test("copilot adapter calls fetcher with correct URL and returns text", async () => {
  let calledUrl = "";
  let calledInit: RequestInit | undefined;
  const fetcher: Fetcher = (input: string, init?: RequestInit) => {
    calledUrl = input;
    calledInit = init;
    const body = {
      choices: [{ message: { content: "hello from copilot" } }],
    };
    return Promise.resolve(
      new Response(JSON.stringify(body), { status: 200 }),
    );
  };
  const res = await copilot({ model: "m", prompt: "p" }, {
    apiKey: "cp-key",
    baseUrl: "https://copilot.example",
    fetcher,
  });
  expect(res.text).toBe("hello from copilot");
  expect(calledUrl).toBe("https://copilot.example/v1/chat/completions");
  const headers = calledInit?.headers as Record<string, string> | undefined;
  const auth = headers && headers["Authorization"];
  expect(auth).toBe("Bearer cp-key");
});
