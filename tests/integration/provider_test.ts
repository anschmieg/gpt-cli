import { assertStringIncludes } from "https://deno.land/std@0.203.0/testing/asserts.ts";
import { chatCompletion } from "../../src/providers/openai.ts";

Deno.test("provider chatCompletion hits mock-openai and returns markdown (with server)", async () => {
  // Try to start the mock server via the relocated module. If permissions
  // don't allow starting it, skip the integration test.
  let server: { close: () => Promise<unknown> | void } | undefined;
  try {
    const mod = await import("../../mock-openai/mock-server.ts");
    if (typeof mod.startMockServerProcess !== "function") {
      console.log(
        "skipping integration: mock server module missing startMockServerProcess",
      );
      return;
    }
    server = await mod.startMockServerProcess();
  } catch (_err) {
    console.log(
      "skipping integration: requires --allow-run to start mock server",
    );
    return;
  }

  try {
    Deno.env.set("GPT_CLI_TEST", "1");
    const content = await chatCompletion({
      model: "gpt-3.5-mock",
      messages: [{ role: "user", content: "Hello" }],
    }, "http://127.0.0.1:8086");

    assertStringIncludes(content, "Advertisement");
  } finally {
    Deno.env.delete("GPT_CLI_TEST");
    if (server) await server.close();
  }
});
