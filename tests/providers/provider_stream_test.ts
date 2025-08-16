import { expect } from "jsr:@std/expect";
import { chatCompletionRequestStream } from "../../src/providers/openai_request.ts";

// Helper: create a Response-like object with a ReadableStream body that emits SSE-style chunks
function makeFakeResponse(chunks: string[]) {
  const encoder = new TextEncoder();
  const stream = new ReadableStream<Uint8Array>({
    start(controller) {
      for (const c of chunks) {
        controller.enqueue(encoder.encode(c));
      }
      controller.close();
    },
  });
  return {
    ok: true,
    body: stream,
    status: 200,
    statusText: "OK",
    text: () => Promise.resolve(""),
    json: () => Promise.resolve({}),
  } as unknown as Response;
}

Deno.test("chatCompletionRequestStream yields SSE data chunks", async () => {
  const content = "Hello world from the server. This is streamed.";
  const words = content.split(/\s+/).filter(Boolean);
  const batchSize = 5;
  const chunks: string[] = [];
  for (let i = 0; i < words.length; i += batchSize) {
    const chunk = words.slice(i, i + batchSize).join(" ") +
      (i + batchSize < words.length ? " " : "");
    const payload = JSON.stringify({
      choices: [{ delta: { content: chunk } }],
    });
    chunks.push(`data: ${payload}\n\n`);
  }
  chunks.push("data: [DONE]\n\n");

  const fakeFetcher = () => Promise.resolve(makeFakeResponse(chunks));

  const gen = chatCompletionRequestStream({
    url: "http://example",
    apiKey: "key",
    body: { messages: [], stream: true },
    fetcher: fakeFetcher,
  });
  const parts: string[] = [];
  for await (const p of gen) {
    parts.push(p);
  }
  const assembled = parts.join("");
  expect(assembled).toBe(content);
});
