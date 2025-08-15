import { runCli } from "../src/cli.ts";

Deno.test("runCli returns greeting", async () => {
  const out = await runCli();
  if (typeof out !== "string") throw new Error("output not a string");
  if (!out.includes("Hello from Deno CLI")) {
    throw new Error(`unexpected output: ${out}`);
  }
});

Deno.test("runCli echoes args", async () => {
  const out = await runCli(["--prompt", "hello"]);
  if (!out.includes("Args:")) throw new Error("did not echo args");
});
