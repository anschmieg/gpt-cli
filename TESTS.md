# Tests for gpt-cli

This document explains the project's test strategy and how to run tests locally.

## Test types

- Unit tests: fast, deterministic, no network or process permissions required. They use dependency injection to mock network calls.
  - Location: `tests/*_test.ts` (e.g. `tests/provider_unit_test.ts`, `tests/utils_test.ts`)
- Integration tests: exercise the full stack by starting the local mock OpenAI server and calling it over HTTP.
  - Location: `tests/integration/*_test.ts` (e.g. `tests/integration/provider_integration_test.ts`)

## Run unit tests locally

Unit tests do not require special permissions:

```bash
deno test tests/cli_test.ts tests/config_test.ts tests/provider_unit_test.ts tests/core_test.ts tests/utils_test.ts --allow-read
```

Or run all unit tests with a glob (depending on shell):

```bash
deno test tests/*_test.ts --allow-read
```

## Run integration tests locally

Integration tests spawn the Deno mock server and require extra permissions. Run them with:

```bash
deno test tests/integration/provider_integration_test.ts --allow-run --allow-net=127.0.0.1:8086 --allow-env --allow-read
```

Note: the test will try to start the Deno mock server (`mock-openai/mock-server.ts`) and will fall back to `node` or `bun` if Deno isn't available.

## CI behavior

- Unit tests run on every push and pull request.
- Integration tests run only for tags or published releases (see `.github/workflows/ci.yml`).

## Tips

- To run a single test file with more output, add `--quiet` to reduce noise or omit it to see detailed output.
- If you change the mock server, ensure it still exposes `/health` and `/v1/chat/completions`.
