import { type CallProviderFn, runCore } from "../../core.ts";
import { ProviderError } from "../../src/providers/adapter_utils.ts";
import type { CoreConfig } from "../../core.ts";

Deno.test("runCore retries once when model_not_supported and flag enabled", async () => {
  let calls = 0;
  const fakeCallProvider: CallProviderFn = (_c) => {
    calls++;
    if (calls === 1) {
      // simulate provider rejecting model
      throw new ProviderError(
        "The requested model is not supported.",
        "model_not_supported",
        { foo: 1 },
      );
    }
    return Promise.resolve({ text: "ok after retry" });
  };

  const cfg: CoreConfig = {
    provider: "openai",
    model: "unsupported-model",
    prompt: "Hello",
    useMarkdown: false,
    autoRetryModel: true,
  };

  const res = await runCore(
    cfg,
    fakeCallProvider,
    undefined,
    undefined,
    undefined,
  );
  if (!res || (res as { ok: boolean }).ok !== true) {
    throw new Error("expected runCore to return ok after retry");
  }
});
