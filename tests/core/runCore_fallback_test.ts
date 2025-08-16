import { runCore } from "../../core.ts";
import { mockFetcher } from "../helpers/mock_fetchers.ts";

Deno.test("runCore falls back when stream requested but adapter lacks streaming", async () => {
  // Use the factory-backed openai adapter but provide a fetcher that returns
  // a non-streaming JSON response. runCore should attempt streaming, then
  // fall back to the non-streaming path and succeed.
  const res = await runCore(
    { provider: "openai", stream: true, prompt: "hi" },
    undefined,
    undefined,
    undefined,
    {
      baseUrl: "http://127.0.0.1:8086",
      apiKey: "DUMMY",
      fetcher: mockFetcher({ choices: [{ message: { content: "ok" } }] }),
    },
  );
  if (!res || (res as { ok: boolean }).ok !== true) {
    throw new Error("expected ok fallback");
  }
});
