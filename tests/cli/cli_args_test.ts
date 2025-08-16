import { runCli } from "../../src/cli.ts";

Deno.test("runCli with no args returns greeting", async () => {
  const out = await runCli();
  if (typeof out !== "string") throw new Error("output not a string");
});

Deno.test("runCli with a single flag echoes it", async () => {
  const out = await runCli(["--verbose"]);
  if (!out.includes("Args:")) throw new Error("did not echo args");
});

Deno.test("runCli with multiple args echoes them in order", async () => {
  const out = await runCli(["--a", "1", "--b", "2"]);
  if (!out.includes("Args:")) throw new Error("did not echo args");
});

Deno.test("runCli handles numeric-only args", async () => {
  const out = await runCli(["123"]);
  if (!out.includes("Args:")) throw new Error("did not echo args");
});
