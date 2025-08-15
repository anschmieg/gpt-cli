import { runCore } from "../core.ts";

Deno.test("runCore prints markdown when response.markdown is present", async () => {
  const fakeProvider = () => Promise.resolve({ markdown: "# title" });
  const out: string[] = [];
  const fakeLog = (...args: unknown[]) => out.push(String(args.join(" ")));
  // Capture console.log
  const orig = console.log;
  try {
    console.log = (...args: unknown[]) => {
      out.push(String(args.join(" ")));
    };
    await runCore(
      { verbose: true },
      fakeProvider,
      (s: string) => `MD:${s}`,
      fakeLog,
    );
  } finally {
    console.log = orig;
  }
  const joined = out.join("\n");
  if (!joined.includes("MD:# title")) throw new Error("didn't render markdown");
});

Deno.test("runCore prints text when response.text is present", async () => {
  const fakeProvider = () => Promise.resolve({ text: "plain text" });
  const out: string[] = [];
  const fakeLog = (...args: unknown[]) => out.push(String(args.join(" ")));
  const orig = console.log;
  try {
    console.log = (...args: unknown[]) => {
      out.push(String(args.join(" ")));
    };
    await runCore({ verbose: true }, fakeProvider, undefined, fakeLog);
  } finally {
    console.log = orig;
  }
  const joined = out.join("\n");
  if (!joined.includes("plain text")) throw new Error("didn't print text");
});
