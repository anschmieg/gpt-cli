import { log } from "../utils/log.ts";
import { renderMarkdown } from "../utils/markdown.ts";

Deno.test("renderMarkdown returns input unchanged", () => {
  const input = "# hi";
  const out = renderMarkdown(input);
  if (out !== input) throw new Error("markdown renderer changed input");
});

Deno.test("log prints only when verbose env is set", () => {
  const orig = console.log;
  const captured: string[] = [];
  console.log = (...args: unknown[]) => {
    captured.push(String(args.join(" ")));
  };
  try {
    Deno.env.delete("GPT_CLI_VERBOSE");
    log("x");
    if (captured.length !== 0) {
      throw new Error("log printed when it should not");
    }

    Deno.env.set("GPT_CLI_VERBOSE", "1");
    log("y");
    if (captured.length === 0) {
      throw new Error("log did not print when verbose");
    }
  } finally {
    console.log = orig;
  }
});
