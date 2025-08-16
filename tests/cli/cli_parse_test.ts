import { parseArgs } from "../../src/cli.ts";

Deno.test("parseArgs handles booleans and aliases", () => {
  const parsed = parseArgs(["--verbose", "-v"]);
  if (!parsed.verbose) throw new Error("verbose not parsed");
});

Deno.test("parseArgs reads string flags and aliases (format)", () => {
  const parsed = parseArgs(["--format", "json"]);
  if (parsed.format !== "json") throw new Error("format not parsed");
});

Deno.test("parseArgs returns undefined for missing optional flags", () => {
  const parsed = parseArgs([]);
  if (parsed.format !== undefined) throw new Error("expected undefined");
});

Deno.test("parseArgs coerces non-numeric count to undefined", () => {
  const parsed = parseArgs(["--count", "not-a-number"]);
  if (parsed.count !== undefined) {
    throw new Error("expected undefined for count");
  }
});

Deno.test("parseArgs preserves multiple positional args", () => {
  const parsed = parseArgs(["one", "two", "three"]);
  if (!Array.isArray(parsed.positional) || parsed.positional.length !== 3) {
    throw new Error("positional args missing");
  }
});
