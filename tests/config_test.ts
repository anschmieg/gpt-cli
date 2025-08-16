Deno.test("deno.json contains expected keys (permission-aware)", async () => {
  const perm = await Deno.permissions.query({
    name: "read",
    path: "deno.json",
  });
  if (perm.state !== "granted") {
    console.log("skipping deno.json test: requires --allow-read");
    return;
  }
  const raw = await Deno.readTextFile("deno.json");
  const cfg = JSON.parse(raw);
  if (!cfg.name) throw new Error("deno.json missing 'name'");
  if (!cfg.entry) throw new Error("deno.json missing 'entry'");
});
