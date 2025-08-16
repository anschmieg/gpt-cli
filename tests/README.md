# Tests

This project groups tests into logical subfolders so you can run focused groups
easily:

- `tests/cli/` - CLI unit tests (parsing, small helpers)
- `tests/providers/` - adapter tests and unit tests for the HTTP client (runtime
  adapters live in `adapters/`)
- `tests/integration/` - integration tests that spawn the mock server (require
  `--allow-run` and network to localhost)

# Tests

The test suite is split into logical groups so you can run focused subsets
quickly.

- `tests/cli/` — CLI unit tests (parsing, argument handling, helpers)
- `tests/providers/` — provider adapters and unit tests for the HTTP client (no
  network)
- `tests/integration/` — integration tests that spawn the mock server (local
  network + runner)

Permission model

- Tests that need Deno privileges now check permissions at runtime and will skip
  with a clear message when the required permission is not granted. That means
  `deno test` without flags should not error — some tests will simply be
  skipped.
- Common permissions:
  - `--allow-read` — required by tests that read project files (e.g.,
    `deno.json`).
  - `--allow-env` — required by tests that read or set environment variables for
    provider adapters.
  - `--allow-run` and `--allow-net=127.0.0.1:8086` — required by integration
    tests that spawn the mock server and talk to it.

Quick commands

Run only CLI tests (fast, no network):

```bash
deno test tests/cli --allow-read --no-check
```

Run provider unit tests (no network):

```bash
deno test tests/providers --allow-read --no-check
```

Run provider tests that require environment variables:

```bash
deno test tests/providers --allow-read --allow-env --no-check
```

Run integration tests (spawns mock server locally):

```bash
deno test tests/integration --allow-run --allow-net=127.0.0.1:8086 --allow-env --allow-read --no-check
```

Run the unit set used by CI (no integration server):

```bash
deno test tests/cli tests/providers tests/config_test.ts tests/core_test.ts tests/utils_test.ts --allow-read --no-check
```

Run everything (local only; allows env and running the mock server):

```bash
deno test tests/cli tests/providers tests/config_test.ts tests/core_test.ts tests/utils_test.ts tests/integration --allow-run --allow-net=127.0.0.1:8086 --allow-env --allow-read --no-check
```

CI notes

- The GitHub Actions workflow separates unit and integration jobs. Unit tests
  run on every push/PR; integration tests run only for tags/releases (or when
  explicitly enabled) because they require `--allow-run` and a runnable
  environment.
- If you need to enable integration tests in CI for feature branches, ensure the
  runner has the proper permissions and the mock server can be started (Deno
  must be available on the runner).

Local troubleshooting

- If an integration test fails to start the mock server, confirm the runner
  (Deno/node/bun) is installed and that you passed `--allow-run` and
  `--allow-net=127.0.0.1:8086`.
- If a provider adapter test is skipped, re-run adding `--allow-env` and ensure
  required env vars are set (e.g., `OPENAI_API_KEY`, `COPILOT_API_KEY`,
  `GEMINI_API_KEY`) if you want the env-branch tests to run.

Developer tips

- Tests now intentionally skip when permissions are absent; watch the console
  output for informative skip messages like
  `skipping integration: requires --allow-run to start mock server`.
- To run a single file with verbose output, invoke it directly:

```bash
deno test tests/providers/provider_unit_test.ts --allow-read --no-check
```

- When changing the mock server, keep `/health` and the `/v1/chat/completions`
  contract intact — integration tests rely on the health endpoint to detect
  readiness.

If you'd like, I can add a small `tests/_helpers.ts` helper to centralize
permission checks and skip messaging; this would reduce repeated
permission-check boilerplate across tests.
