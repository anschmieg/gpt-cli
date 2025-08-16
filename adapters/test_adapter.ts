import type { ProviderOptions } from "../src/providers/types.ts";

export function callProvider(_config: unknown, opts?: ProviderOptions) {
  if (!opts?.apiKey) throw new Error("TEST_ADAPTER_API_KEY required");
  return { text: "chunk-1chunk-2" };
}

export async function* chatCompletionStream(
  _body: unknown,
  opts?: ProviderOptions,
) {
  if (!opts?.apiKey) throw new Error("TEST_ADAPTER_API_KEY required");
  yield "chunk-1";
  yield "chunk-2";
}
