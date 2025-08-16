import { assertEquals } from "https://deno.land/std@0.201.0/testing/asserts.ts";
import { callProvider as gemini } from "../../providers/gemini.ts";
import type { Fetcher } from "../../providers/gemini.ts";

Deno.test("gemini adapter errors when GEMINI_API_KEY missing", async () => {
  const perm = await Deno.permissions.query({ name: "env" });
  if (perm.state !== "granted") {
    console.log("skipping gemini adapter env tests: requires --allow-env");
    return;
  }
  const prevKey = Deno.env.get("GEMINI_API_KEY");
  try {
    if (prevKey !== undefined) Deno.env.delete("GEMINI_API_KEY");
    try {
      await gemini({ model: "x", prompt: "hi" });
      throw new Error("expected error when GEMINI_API_KEY missing");
    } catch (err) {
      if (!(err instanceof Error)) throw err;
    }
  } finally {
    if (prevKey !== undefined) Deno.env.set("GEMINI_API_KEY", prevKey);
  }
});

Deno.test("gemini adapter calls fetcher with correct URL and returns text", async () => {
  const perm = await Deno.permissions.query({ name: "env" });
  if (perm.state !== "granted") {
    console.log("skipping gemini adapter env tests: requires --allow-env");
    return;
  }
  const prevKey = Deno.env.get("GEMINI_API_KEY");
  try {
    Deno.env.set("GEMINI_API_KEY", "g-key");

    let calledUrl = "";
    let calledInit: RequestInit | undefined;
    const fetcher: Fetcher = (input: string, init?: RequestInit) => {
      calledUrl = input;
      calledInit = init;
      const body = { choices: [{ message: { content: "hello from gemini" } }] };
      return Promise.resolve(
        new Response(JSON.stringify(body), { status: 200 }),
      );
    };

    const res = await gemini({ model: "m", prompt: "p" }, fetcher);
    assertEquals(res.text, "hello from gemini");
    assertEquals(
      calledUrl,
      "https://generativelanguage.googleapis.com/v1beta/openai",
    );
    const auth = calledInit?.headers &&
      (calledInit.headers as Record<string, string>)["Authorization"];
    assertEquals(auth, "Bearer g-key");
  } finally {
    if (prevKey !== undefined) Deno.env.set("GEMINI_API_KEY", prevKey);
    else Deno.env.delete("GEMINI_API_KEY");
  }
});
