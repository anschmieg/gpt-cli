import { parseArgs } from "../src/cli.ts";
import { assertEquals } from "https://deno.land/std@0.201.0/testing/asserts.ts";

Deno.test("parseArgs handles booleans and aliases", () => {
  const res = parseArgs(["--help", "-v"]);
  assertEquals(res.help, true);
  assertEquals(res.verbose, true);
  assertEquals(res.positional, []);
});

Deno.test("parseArgs reads string flags and aliases", () => {
  const res = parseArgs(["--format", "json", "-c", "5", "pos1"]);
  assertEquals(res.format, "json");
  assertEquals(res.count, 5);
  assertEquals(res.positional, ["pos1"]);
});

Deno.test("parseArgs returns undefined for missing optional flags", () => {
  const res = parseArgs(["arg1"]);
  assertEquals(res.format, undefined);
  assertEquals(res.count, undefined);
  assertEquals(res.positional, ["arg1"]);
});

Deno.test("parseArgs coerces non-numeric count to undefined", () => {
  const res = parseArgs(["-c", "not-a-number"]);
  assertEquals(res.count, undefined);
});

Deno.test("parseArgs preserves multiple positional args", () => {
  const res = parseArgs(["one", "two", "three"]);
  assertEquals(res.positional, ["one", "two", "three"]);
});
