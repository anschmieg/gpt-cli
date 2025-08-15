import { runCli } from "../src/cli.ts";
import { assertEquals } from "https://deno.land/std@0.201.0/testing/asserts.ts";

Deno.test("runCli handles empty-string arg", async () => {
  const args = [""];
  const out = await runCli(args);
  const expected = `Hello from Deno CLI! Args: ${args.join(" ")}`;
  assertEquals(out, expected);
});

Deno.test("runCli handles whitespace-only args", async () => {
  const args = [" ", "   "];
  const out = await runCli(args);
  const expected = `Hello from Deno CLI! Args: ${args.join(" ")}`;
  assertEquals(out, expected);
});

Deno.test("runCli with a very long single arg", async () => {
  const long = "a".repeat(10_000);
  const args = [long];
  const out = await runCli(args);
  const expected = `Hello from Deno CLI! Args: ${args.join(" ")}`;
  assertEquals(out, expected);
});

Deno.test("runCli with many args (stress, moderate)", async () => {
  const args = Array.from({ length: 500 }, (_, i) => `x${i}`);
  const out = await runCli(args);
  const expected = `Hello from Deno CLI! Args: ${args.join(" ")}`;
  assertEquals(out, expected);
});

Deno.test("runCli preserves Unicode and combining characters", async () => {
  const args = ["你好", "🌟✨", "\u0065\u0301"];
  const out = await runCli(args);
  const expected = `Hello from Deno CLI! Args: ${args.join(" ")}`;
  assertEquals(out, expected);
});

Deno.test("runCli preserves newlines and NUL characters inside args", async () => {
  const args = ["line1\nline2", "contain\0null"];
  const out = await runCli(args);
  const expected = `Hello from Deno CLI! Args: ${args.join(" ")}`;
  assertEquals(out, expected);
});
