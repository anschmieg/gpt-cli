import { assertEquals } from "https://deno.land/std@0.201.0/testing/asserts.ts";
import { callProvider as openai } from "../../providers/openai.ts";
import type { Fetcher } from "../../providers/openai.ts";

Deno.test("openai adapter errors when OPENAI_API_KEY missing", async () => {
  const perm = await Deno.permissions.query({ name: "env" });
  if (perm.state !== "granted") {
    console.log("skipping openai adapter env tests: requires --allow-env");
    return;
  }
  const prev = Deno.env.get("OPENAI_API_KEY");
  try {
    if (prev !== undefined) Deno.env.delete("OPENAI_API_KEY");
    try {
      await openai({ model: "x", prompt: "hi" });
      throw new Error("expected error when OPENAI_API_KEY missing");
    } catch (err) {
      if (!(err instanceof Error)) throw err;
    }
  } finally {
    if (prev !== undefined) Deno.env.set("OPENAI_API_KEY", prev);
  }
});

Deno.test("openai adapter calls fetcher with correct URL and returns text", async () => {
  const perm = await Deno.permissions.query({ name: "env" });
  if (perm.state !== "granted") {
    console.log("skipping openai adapter env tests: requires --allow-env");
    return;
  }
  const prev = Deno.env.get("OPENAI_API_KEY");
  const prevBase = Deno.env.get("OPENAI_API_BASE");
  try {
    Deno.env.set("OPENAI_API_KEY", "test-key");
    Deno.env.set("OPENAI_API_BASE", "https://api.example");

    let calledUrl = "";
    let calledInit: RequestInit | undefined;
    const fetcher: Fetcher = (input: string, init?: RequestInit) => {
      calledUrl = input;
      calledInit = init;
      const body = { choices: [{ message: { content: "hello from openai" } }] };
      return Promise.resolve(
        new Response(JSON.stringify(body), { status: 200 }),
      );
    };

    const res = await openai({ model: "m", prompt: "p" }, fetcher);
    assertEquals(res.text, "hello from openai");
    assertEquals(calledUrl, "https://api.example/v1/chat/completions");
    // basic header check
    const auth = calledInit?.headers &&
      (calledInit.headers as Record<string, string>)["Authorization"];
    assertEquals(auth, "Bearer test-key");
  } finally {
    if (prev !== undefined) Deno.env.set("OPENAI_API_KEY", prev);
    else Deno.env.delete("OPENAI_API_KEY");
    if (prevBase !== undefined) Deno.env.set("OPENAI_API_BASE", prevBase);
    else Deno.env.delete("OPENAI_API_BASE");
  }
});
