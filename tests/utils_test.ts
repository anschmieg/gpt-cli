import { debug } from "../src/utils/log.ts";
import { renderMarkdown } from "../src/utils/markdown.ts";

Deno.test("renderMarkdown returns input unchanged", () => {
  // Plain text: should be unchanged
  const plain = "hello world";
  const outPlain = renderMarkdown(plain);
  if (outPlain !== plain) {
    throw new Error("markdown renderer changed plain text");
  }

  // Markdown: should be rendered
  const md = "# hi";
  const outMd = renderMarkdown(md);
  if (outMd === md) {
    throw new Error("markdown renderer did not render markdown");
  }
});

Deno.test("log prints only when verbose env is set", async () => {
  const perm = await Deno.permissions.query({ name: "env" });
  if (perm.state !== "granted") {
    console.log("skipping utils env tests: requires --allow-env");
    return;
  }
  const orig = console.log;
  const captured: string[] = [];
  console.log = (...args: unknown[]) => {
    captured.push(String(args.join(" ")));
  };
  try {
    Deno.env.delete("GPT_CLI_VERBOSE");
    debug("x");
    if (captured.length !== 0) {
      throw new Error("log printed when it should not");
    }

    Deno.env.set("GPT_CLI_VERBOSE", "1");
    debug("y");
    if (captured.length === 0) {
      throw new Error("log did not print when verbose");
    }
  } finally {
    console.log = orig;
  }
});
