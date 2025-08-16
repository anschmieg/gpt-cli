#!/usr/bin/env -S deno run --allow-read --allow-run --allow-env
// Deno TypeScript pre-commit script
// - runs deno fmt, deno lint, and the full unit test suite
// - stages formatted tracked files (`git add -u`)
// - prints a compact boxed layout using Unicode and proper string width handling

// prefer simple Unicode symbols for terminal compatibility
const OK = "✓";
const WARN = "⚠";
const FAIL = "✖";
const YELLOW = "\x1b[33m";
const RED = "\x1b[31m";

const CYAN = "\x1b[36m";
const GREEN = "\x1b[32m";
const BOLD = "\x1b[1m";
const RESET = "\x1b[0m";

import boxen from "https://esm.sh/boxen@8.0.1";

// using `stringWidth(stripAnsi(...))` directly where needed
// simplified: rely on boxen/text wrapping for long filenames; no truncation helper

async function runCmd(cmd: string[], label: string) {
  const command = new Deno.Command(cmd[0], {
    args: cmd.slice(1),
    stdout: "piped",
    stderr: "piped",
  });
  const output = await command.output();
  const out = new TextDecoder().decode(output.stdout) +
    new TextDecoder().decode(output.stderr);
  if (!output.success) {
    // print errors in bold red
    console.error(
      `\n${RED}${BOLD}${FAIL} ${label} failed — aborting commit${RESET}\n`,
    );
    console.error(
      `${BOLD}--------------------------------------------------${RESET}`,
    );
    console.error(`${BOLD}${out}${RESET}`);
    console.error(
      `${BOLD}--------------------------------------------------${RESET}`,
    );
    Deno.exit(1);
  }
  return out;
}

function renderStatusLabel(label: string, level: "ok" | "warn" | "fail") {
  const color = level === "ok" ? GREEN : level === "warn" ? YELLOW : RED;
  const sym = level === "ok" ? OK : level === "warn" ? WARN : FAIL;
  const boldStart = level === "fail" ? BOLD : "";
  return `${color}${boldStart}${sym} ${label}${RESET}`;
}

async function main() {
  const statuses: Record<string, "ok" | "warn" | "fail"> = {};

  await runCmd(["deno", "fmt"], "deno fmt");
  statuses["fmt"] = "ok";

  // stage formatted tracked files
  try {
    await runCmd(["git", "add", "-u"], "git add -u");
    statuses["stg"] = "ok";
  } catch {
    statuses["stg"] = "warn";
  }

  const stagedOut = await runCmd(
    ["git", "diff", "--cached", "--name-only"],
    "git diff --cached",
  )
    .catch(() => "");
  const staged = stagedOut.trim().split(/\r?\n/).filter(Boolean);

  await runCmd(["deno", "lint"], "deno lint");
  statuses["lint"] = "ok";

  await runCmd(["deno", "task", "test:unit"], "deno task test:unit");
  statuses["tests"] = "ok";

  // build status line (use full words instead of command stubs)
  const statusLine = `${renderStatusLabel("format", statuses["fmt"])}   ${
    renderStatusLabel("stage", statuses["stg"])
  }   ${renderStatusLabel("lint", statuses["lint"])}   ${
    renderStatusLabel("tests", statuses["tests"])
  }`;

  // build box content using boxen for reliable width handling
  let body = `${BOLD}Pre-commit Checks${RESET}\n`;
  body += statusLine;
  if (staged.length > 0) {
    body += `\n\n${BOLD}Staged files (${staged.length})${RESET}\n`;
    for (const f of staged) {
      body += `${CYAN}${f}${RESET}\n`;
    }
  }

  const boxed = boxen(body.trim(), {
    padding: 1,
    borderColor: "cyan",
    borderStyle: "round",
  });
  console.log(boxed);
  console.log(`${GREEN}${BOLD}${OK} Ready for commit\n${RESET}`);
}

if (import.meta.main) main();
