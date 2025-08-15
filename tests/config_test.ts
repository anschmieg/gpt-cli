Deno.test("deno.json contains expected keys", async () => {
  const raw = await Deno.readTextFile("deno.json");
  const cfg = JSON.parse(raw);
  if (!cfg.name) throw new Error("deno.json missing 'name'");
  if (!cfg.entry) throw new Error("deno.json missing 'entry'");
});
