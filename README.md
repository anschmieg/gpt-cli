# gpt-cli

A small, test-friendly Deno CLI that acts as a thin wrapper around provider
APIs (OpenAI-compatible by default) and contains a fast local mock server used
by the integration tests.

This README documents how the program is wired, what the runtime defaults are,
how to run the CLI for development, and how to run the unit and integration
test suites used by CI.

## What the program does

- `./cli.ts` is the real CLI entry used when running the program normally. It
	parses full user-facing options and calls `runCore` which orchestrates the
	request to a provider and renders output.
- `./src/cli.ts` contains a small, testable `parseArgs` helper and a `runCli`
	shim used by unit tests. `runCli` currently returns a greeting and echoes
	provided arguments (keeps tests fast and deterministic).
- A Deno-based mock OpenAI-compatible server lives under `mock-openai/` and is
	used by the integration tests to avoid network calls to external APIs.

## Defaults

- provider: `copilot` (default when running the top-level `cli.ts`)
- model: `gpt-4.1-mini` (default model)
- temperature: `0.6` (default)
- verbose: `false`
- system: the default system prompt below (used if `--system` not provided)
- file: (none by default; set with `--file`)
- markdown output: enabled by default. When disabled (`--no-markdown` or
	`--markdown=false`), the CLI will output plain text only.

Default system prompt:

```
You are an AI assistant called via CLI. Respond concisely and clearly, focusing only on the user's prompt. Include only very brief explanations unless explicitly asked.
```

When markdown is enabled, providers' markdown should be returned as-is and the
CLI will output it without adding any top-level wrapper. When markdown is
disabled the CLI will output plain text only (prefer `text` fields from the
provider response and fall back to markdown if no text is available).

The testing-oriented `parseArgs` in `src/cli.ts` supports the following flags
and defaults (intended for unit tests and quick parsing):

- `--help`, `-h` (boolean)
- `--verbose`, `-v` (boolean; default: `false`)
- `--format`, `-f` (string)
- `--count`, `-c` (string -> coerced to number when numeric)
- Positional arguments are returned as `positional` in the parse result.

Note: The top-level `cli.ts` provides the real user-facing flags such as
`--provider`, `--model`, `--temperature`, `--system`, and `--file`.

## Usage

Run the CLI (development):

```bash
deno run --allow-read main.ts
```

If you run the real CLI and it needs to call provider APIs, you will also need
to grant network and environment permissions depending on your provider setup
(for example, `--allow-net` and `--allow-env`). Unit tests and the `src` test
helpers don't require network permissions.

## Tests

Unit tests are fast and do not require network access. Integration tests start
the local mock OpenAI server and therefore require a few Deno permissions.

Run unit tests only:

```bash
deno test --allow-read
```

Run the full test suite (includes integration tests that spawn the mock
server):

```bash
deno test --allow-run --allow-net=127.0.0.1:8086 --allow-env --allow-read
```

Integration tests set the `GPT_CLI_TEST=1` environment variable when they
spawn the mock server; this is a safety flag used by the provider code to
prevent accidental network calls to non-local endpoints during tests.

## CI behavior

The repository includes a GitHub Actions workflow (`.github/workflows/ci.yml`) that:

- Runs `deno lint` and unit tests on `push` and `pull_request` events.
- Runs integration tests only on tags or release events (they spawn the
	mock server and require network/run permissions in CI).
- On release/tag events, CI also runs tests with coverage and uploads an
	`lcov` artifact.

## Coverage

The project can produce an lcov report via Deno's built-in coverage tooling:

```bash
deno test --coverage=coverage --allow-run --allow-net=127.0.0.1:8086 --allow-env --allow-read
deno coverage --lcov coverage > coverage.lcov
```

## Mock server

The mock server is implemented in `mock-openai/` and is intentionally local
and lightweight so integration tests can run offline and deterministically.

## Notes for contributors

- Run `deno fmt` before committing to keep formatting consistent.
- `deno lint` is run in CI; fix any lint issues locally before opening a PR.
- If you add tests that spawn processes or require network, follow the
	existing patterns and gate those tests behind the `GPT_CLI_TEST` guard where
	appropriate.

## License

This project is MIT-licensed. See `LICENSE` for details (if present).

