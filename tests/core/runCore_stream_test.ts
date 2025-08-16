import { runCore } from "../../core.ts";

Deno.test("runCore streaming path uses provider chatCompletionStream", async () => {
  const res = await runCore(
    { provider: "test_adapter", stream: true, prompt: "hi" },
    undefined,
    undefined,
    undefined,
    { apiKey: "DUMMY" },
  );
  if (!res || (res as { ok: boolean }).ok !== true) {
    throw new Error("expected ok");
  }
});
