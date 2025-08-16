#!/usr/bin/env -S deno run --allow-read
import { validateAdapterModule } from "../src/providers/adapter_utils.ts";

const adapters = ["openai", "copilot", "gemini"];

for (const name of adapters) {
  try {
    const m = await import(`../providers/${name}.ts`);
    validateAdapterModule(m, name);
    console.log(name, "OK");
  } catch (err) {
    const obj = err as unknown;
    const msg =
      (obj && typeof obj === "object" &&
          (obj as Record<string, unknown>)["message"])
        ? String((obj as Record<string, unknown>)["message"])
        : String(err);
    console.error(name, "FAILED:", msg);
    Deno.exit(2);
  }
}
console.log("All adapters OK");
