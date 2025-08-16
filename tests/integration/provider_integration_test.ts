import { assertStringIncludes } from "https://deno.land/std@0.203.0/testing/asserts.ts";

Deno.test("end-to-end streaming from mock server (process)", async () => {
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
    const prov = await import("../../adapters/openai.ts");
    const gen = await prov.chatCompletionStream({
      model: "gpt-3.5-mock",
      messages: [{ role: "user", content: "Hello" }],
      stream: true,
    }, { baseUrl: "http://127.0.0.1:8086", apiKey: "test" });

    let out = "";
    for await (const chunk of gen) {
      out += chunk;
    }

    assertStringIncludes(out, "Advertisement");
  } finally {
    if (server) await server.close();
  }
});
