import { runCli } from "../src/cli.ts";
import { assertEquals } from "https://deno.land/std@0.201.0/testing/asserts.ts";

Deno.test("runCli with no args returns greeting", async () => {
  const out = await runCli([]);
  assertEquals(out, "Hello from Deno CLI!");
});

Deno.test("runCli with a single flag echoes it", async () => {
  const out = await runCli(["--help"]);
  assertEquals(out, "Hello from Deno CLI! Args: --help");
});

Deno.test("runCli with multiple args echoes them in order", async () => {
  const out = await runCli(["prompt", "--verbose", "--count", "3"]);
  assertEquals(out, "Hello from Deno CLI! Args: prompt --verbose --count 3");
});

Deno.test("runCli preserves positional argument spacing when provided as separate entries", async () => {
  const out = await runCli(["multi", "word", "prompt"]);
  assertEquals(out, "Hello from Deno CLI! Args: multi word prompt");
});

Deno.test("runCli handles numeric-only args", async () => {
  const out = await runCli(["123", "456"]);
  assertEquals(out, "Hello from Deno CLI! Args: 123 456");
});
