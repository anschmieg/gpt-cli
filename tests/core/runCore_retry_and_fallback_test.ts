import { expect } from "jsr:@std/expect";
import { type CallProviderFn, runCore } from "../../core.ts";

Deno.test("runCore retries once when model_not_supported and autoRetryModel=true then fails", async () => {
  // First call throws a provider-shaped error indicating model not supported
  let calls = 0;
  const failingProvider: CallProviderFn = (_cfg) => {
    calls++;
    if (calls === 1) {
      // simulate provider error shape
      return Promise.reject({
        error: {
          message: "Requested model is not supported",
          code: "model_not_supported",
        },
      });
    }
    // second call also fails for test purposes
    return Promise.reject(new Error("still failing"));
  };

  const origConsoleError = console.error;
  const errors: string[] = [];
  try {
    console.error = (...args: unknown[]) => errors.push(String(args.join(" ")));
    const res = await runCore({ autoRetryModel: true }, failingProvider);
    expect(res.ok).toBe(false);
    // ensure retry occurred (two calls)
    expect(calls).toBe(2);
    // the reported error message should be from the final failure
    expect(errors.join(" ")).toContain("still failing");
  } finally {
    console.error = origConsoleError;
  }
});

Deno.test("runCore when useMarkdown=false prefers text but falls back to markdown", async () => {
  const provider: CallProviderFn = () =>
    Promise.resolve({ markdown: "# hi", text: undefined });
  const out: string[] = [];
  const origLog = console.log;
  try {
    console.log = (...args: unknown[]) => out.push(String(args.join(" ")));
    const res = await runCore({ useMarkdown: false }, provider);
    expect(res.ok).toBe(true);
    // when text is absent, markdown should be printed
    expect(out.join("\n")).toContain("# hi");
  } finally {
    console.log = origLog;
  }
});
