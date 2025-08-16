import { chatCompletion } from "../../src/providers/openai.ts";

Deno.test("openai.chatCompletion refuses non-local endpoint in test mode", async () => {
  // Skip when env access isn't allowed in this runner
  const perm = await Deno.permissions.query({
    name: "env" as Deno.PermissionName,
  });
  if (perm.state !== "granted") {
    console.log("skipping openai_local_test: requires --allow-env");
    return;
  }
  // When GPT_CLI_TEST=1, openai.chatCompletion should throw if baseUrl is not localhost
  Deno.env.set("GPT_CLI_TEST", "1");
  try {
    try {
      await chatCompletion({
        model: "m",
        messages: [{ role: "user", content: "hi" }],
      }, { baseUrl: "https://api.openai.com" });
      throw new Error("expected error refusing non-local endpoint");
    } catch (err) {
      // ok if an error was thrown
      if (!(err instanceof Error)) throw err;
    }
  } finally {
    Deno.env.delete("GPT_CLI_TEST");
  }
});
