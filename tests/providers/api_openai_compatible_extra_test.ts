import { expect } from "jsr:@std/expect";
import {
  chatCompletionRequest,
  chatCompletionRequestStream,
} from "../../src/providers/openai_request.ts";

// Helper to make a Response-like object with a ReadableStream body
function makeFakeResponseFromChunks(chunks: string[]) {
  const encoder = new TextEncoder();
  const stream = new ReadableStream<Uint8Array>({
    start(controller) {
      for (const c of chunks) controller.enqueue(encoder.encode(c));
      controller.close();
    },
  });
  return {
    ok: true,
    body: stream,
    status: 200,
    statusText: "OK",
    text: () => Promise.resolve("") as Promise<string>,
    json: () => Promise.resolve({}) as Promise<unknown>,
  } as unknown as Response;
}

Deno.test("chatCompletionRequest throws on invalid response shape", async () => {
  const fakeFetcher = () =>
    Promise.resolve({
      ok: true,
      json: () => Promise.resolve({ choices: [{ message: { content: 123 } }] }),
      text: () => Promise.resolve(JSON.stringify({})),
      status: 200,
      statusText: "OK",
    } as unknown as Response);

  let threw = false;
  try {
    await chatCompletionRequest({
      url: "http://example",
      apiKey: "x",
      body: { messages: [] },
      fetcher: fakeFetcher,
    });
  } catch (err) {
    threw = true;
    expect(String(err)).toContain("invalid response shape");
  }
  expect(threw).toBe(true);
});

Deno.test("chatCompletionRequestStream returns no chunks when body has no reader", async () => {
  const fakeFetcher = () =>
    Promise.resolve({
      ok: true,
      body: undefined,
      status: 200,
      statusText: "OK",
      text: () => Promise.resolve(""),
      json: () => Promise.resolve({}),
    } as unknown as Response);

  const gen = await chatCompletionRequestStream({
    url: "http://x",
    apiKey: "k",
    body: { messages: [], stream: true },
    fetcher: fakeFetcher,
  });
  const parts: string[] = [];
  for await (const p of gen) parts.push(p);
  expect(parts.length).toBe(0);
});

Deno.test("chatCompletionRequestStream flushes remaining buffer and ignores parse errors", async () => {
  // First chunk is a valid data packet with trailing \n\n
  const payload1 = JSON.stringify({
    choices: [{ delta: { content: "hello " } }],
  });
  // Second chunk is valid JSON but missing trailing separator to force flush path
  const payload2 = JSON.stringify({
    choices: [{ delta: { content: "world" } }],
  });
  // Also include a non-json data packet to ensure parse errors are ignored
  const chunks = [
    `data: ${payload1}\n\n`,
    `data: not-a-json\n\n`,
    `data: ${payload2}`,
  ];

  const fakeFetcher = () => Promise.resolve(makeFakeResponseFromChunks(chunks));

  const gen = await chatCompletionRequestStream({
    url: "http://x",
    apiKey: "k",
    body: { messages: [], stream: true },
    fetcher: fakeFetcher,
  });
  const parts: string[] = [];
  for await (const p of gen) parts.push(p);
  // should receive hello and world
  expect(parts.join("")).toBe("hello world");
});
