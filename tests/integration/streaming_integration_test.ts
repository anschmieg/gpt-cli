import { assertStringIncludes } from "https://deno.land/std@0.203.0/testing/asserts.ts";

// This integration test simulates the mock server using an in-memory fetcher
// that returns SSE chunks. This avoids needing --allow-run or --allow-net.
Deno.test("end-to-end streaming from mock server (in-memory)", async () => {
  // Import the openai provider adapter and call its streaming API directly.
  const prov = await import("../../adapters/openai.ts");

  // Simulate the same content as the mock server (contains 'Advertisement')
  const content =
    `__Advertisement :)__\n\nThis is a short test response containing Advertisement.`;

  // Create a ReadableStream that yields a few SSE 'data:' packets with JSON payloads
  const encoder = new TextEncoder();
  const s = new ReadableStream({
    start(controller) {
      const words = content.split(/\s+/).filter(Boolean);
      const batchSize = 4;
      for (let i = 0; i < words.length; i += batchSize) {
        const chunk = words.slice(i, i + batchSize).join(" ") +
          (i + batchSize < words.length ? " " : "");
        const payload = JSON.stringify({
          choices: [{ delta: { content: chunk } }],
        });
        controller.enqueue(encoder.encode(`data: ${payload}\n\n`));
      }
      controller.enqueue(encoder.encode("data: [DONE]\n\n"));
      controller.close();
    },
  });

  // Simple fetcher that returns the SSE stream.
  const fetcher = (_input: string, _init?: RequestInit) => {
    return Promise.resolve(
      new Response(s, {
        status: 200,
        headers: { "Content-Type": "text/event-stream" },
      }),
    );
  };

  const gen = await prov.chatCompletionStream({
    model: "gpt-3.5-mock",
    messages: [{ role: "user", content: "Hello" }],
    stream: true,
  }, { fetcher, apiKey: "fake-test-key" });

  let out = "";
  for await (const chunk of gen) {
    out += chunk;
  }

  // The simulated stream contains the word 'Advertisement'
  assertStringIncludes(out, "Advertisement");
});
