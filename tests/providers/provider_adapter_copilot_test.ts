import { assertEquals } from "https://deno.land/std@0.201.0/testing/asserts.ts";
import { callProvider as copilot } from "../../providers/copilot.ts";
import type { Fetcher } from "../../providers/copilot.ts";

Deno.test("copilot adapter errors when COPILOT_API_KEY or BASE missing", async () => {
  const perm = await Deno.permissions.query({ name: "env" });
  if (perm.state !== "granted") {
    console.log("skipping copilot adapter env tests: requires --allow-env");
    return;
  }
  const prevKey = Deno.env.get("COPILOT_API_KEY");
  const prevBase = Deno.env.get("COPILOT_API_BASE");
  try {
    if (prevKey !== undefined) Deno.env.delete("COPILOT_API_KEY");
    if (prevBase !== undefined) Deno.env.delete("COPILOT_API_BASE");
    try {
      await copilot({ model: "x", prompt: "hi" });
      throw new Error("expected error when COPILOT_API_KEY missing");
    } catch (err) {
      if (!(err instanceof Error)) throw err;
    }
  } finally {
    if (prevKey !== undefined) Deno.env.set("COPILOT_API_KEY", prevKey);
    if (prevBase !== undefined) Deno.env.set("COPILOT_API_BASE", prevBase);
  }
});

Deno.test("copilot adapter calls fetcher with correct URL and returns text", async () => {
  const perm = await Deno.permissions.query({ name: "env" });
  if (perm.state !== "granted") {
    console.log("skipping copilot adapter env tests: requires --allow-env");
    return;
  }
  const prevKey = Deno.env.get("COPILOT_API_KEY");
  const prevBase = Deno.env.get("COPILOT_API_BASE");
  try {
    Deno.env.set("COPILOT_API_KEY", "c-key");
    Deno.env.set("COPILOT_API_BASE", "https://copilot.example");

    let calledUrl = "";
    let calledInit: RequestInit | undefined;
    const fetcher: Fetcher = (input: string, init?: RequestInit) => {
      calledUrl = input;
      calledInit = init;
      const body = {
        choices: [{ message: { content: "hello from copilot" } }],
      };
      return Promise.resolve(
        new Response(JSON.stringify(body), { status: 200 }),
      );
    };

    const res = await copilot({ model: "m", prompt: "p" }, fetcher);
    assertEquals(res.text, "hello from copilot");
    assertEquals(calledUrl, "https://copilot.example/v1/chat/completions");
    const auth = calledInit?.headers &&
      (calledInit.headers as Record<string, string>)["Authorization"];
    assertEquals(auth, "Bearer c-key");
  } finally {
    if (prevKey !== undefined) Deno.env.set("COPILOT_API_KEY", prevKey);
    else Deno.env.delete("COPILOT_API_KEY");
    if (prevBase !== undefined) Deno.env.set("COPILOT_API_BASE", prevBase);
    else Deno.env.delete("COPILOT_API_BASE");
  }
});
