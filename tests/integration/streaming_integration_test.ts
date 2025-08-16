import { assertStringIncludes } from "https://deno.land/std@0.203.0/testing/asserts.ts";
import { startMockServerProcess } from "../../mock-openai/mock-server.ts";

// This integration test starts the mock server and calls the core run with streaming enabled.
Deno.test("end-to-end streaming from mock server", async () => {
  // Ensure tests run in test mode
  Deno.env.set("GPT_CLI_TEST", "1");

  const server = await startMockServerProcess();
  try {
    // Import the openai provider adapter and call its streaming API directly.
    const prov = await import("../../providers/openai.ts");
    const gen = await prov.chatCompletionStream({
      model: "gpt-3.5-mock",
      messages: [{ role: "user", content: "Hello" }],
      stream: true,
    }, { baseUrl: "http://127.0.0.1:8086" });

    let out = "";
    for await (const chunk of gen) {
      out += chunk;
    }

    // The mock server returns a short assistant reply containing 'Advertisement' or similar
    assertStringIncludes(out, "Advertisement");
  } finally {
    await server.close();
  }
});
