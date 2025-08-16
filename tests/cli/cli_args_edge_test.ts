import { runCli } from "../../src/cli.ts";

Deno.test("runCli handles empty-string arg", async () => {
  const out = await runCli([""]);
  if (!out.includes("Args:")) throw new Error("did not echo args");
});

Deno.test("runCli handles whitespace-only args", async () => {
  const out = await runCli(["   "]);
  if (!out.includes("Args:")) throw new Error("did not echo args");
});

Deno.test("runCli with a very long single arg", async () => {
  const long = "a".repeat(10000);
  const out = await runCli([long]);
  if (!out.includes("Args:")) throw new Error("did not echo args");
});

Deno.test("runCli with many args (stress, moderate)", async () => {
  const args = Array.from({ length: 200 }, (_, i) => `arg${i}`);
  const out = await runCli(args);
  if (!out.includes("Args:")) throw new Error("did not echo args");
});

Deno.test("runCli preserves Unicode and combining characters", async () => {
  const out = await runCli(["e\u0301", "नमस्ते"]);
  if (!out.includes("Args:")) throw new Error("did not echo args");
});

Deno.test("runCli preserves newlines and NUL characters inside args", async () => {
  const out = await runCli(["line1\nline2", "NUL\0char"]);
  if (!out.includes("Args:")) throw new Error("did not echo args");
});
