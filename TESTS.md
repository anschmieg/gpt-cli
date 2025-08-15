## Unit Tests

Unit tests do not require special permissions:

```bash
deno test tests/cli_test.ts tests/config_test.ts tests/provider_unit_test.ts tests/core_test.ts tests/utils_test.ts --allow-read
```

Or run all unit tests with a glob (depending on shell):

```bash
deno test tests/*_test.ts --allow-read
```

## Run integration tests locally

Integration tests spawn the mock OpenAI-compatible server and require extra permissions. Run them with:

```bash
deno test tests/integration/provider_integration_test.ts --allow-run --allow-net=127.0.0.1:8086 --allow-env --allow-read
```

## CI behavior

- Unit tests run on every push and pull request.
- Integration tests run only for tags or published releases (see `.github/workflows/ci.yml`).

## Tips

- To run a single test file with more output, add `--quiet` to reduce noise or omit it to see detailed output.
- If you change the mock server, ensure it still exposes `/health` and `/v1/chat/completions`.

## CLI behavior and tested expectations

This section documents the exact runtime behavior the CLI should provide and which tests assert each behavior.

- Parsing and help
	- Expected behavior: when run with `-h` or no prompt argument, the CLI prints help and exits with code 0.
	- Verified by: `tests/cli_test.ts` which imports `parseArgs` and checks that help-related flags are present and that the parsed config contains the prompt when provided.

- Configuration mapping
	- Expected behavior: parsed CLI options (`--provider`, `--model`, `--temperature`, `--system`, `--file`, `--verbose`) map into a config object passed to core.
	- Verified by: `tests/cli_test.ts` (parsing assertions) and `tests/core_test.ts` which exercises `runCore` with explicit configs.

- Core control flow
	- Expected behavior: `runCore` calls the provider, logs debug output when `verbose` is true, formats markdown output via `renderMarkdown` when provider returns `markdown`, otherwise prints `text`.
	- Verified by: `tests/core_test.ts` which injects a fake provider and a fake renderer and asserts console output includes rendered markdown or plaintext.

- Provider isolation
	- Expected behavior: provider modules should support dependency injection for the HTTP client so unit tests can run without network; in test mode (`GPT_CLI_TEST=1`) providers refuse non-local endpoints.
	- Verified by: `tests/provider_unit_test.ts` (injects fake fetcher to test success and error handling) and `tests/integration/provider_integration_test.ts` (integration against the Deno mock server).

- Utilities
	- Expected behavior: `renderMarkdown` returns text (placeholder for future formatting), and `log` only prints when `GPT_CLI_VERBOSE` is `1`.
	- Verified by: `tests/utils_test.ts`.

If you'd like, I can expand this section into a living checklist (e.g., YAML or JSON) that maps tests to behaviors and includes links to code locations for faster navigation.
