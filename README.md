gpt-cli — Go rewrite (Bubble Tea TUI planned)

This repository is a Go rewrite of the original gpt-cli project. The focus of the
current branch is to implement a small, well-tested core that supports streaming
provider responses and safe incremental markdown rendering. A Bubble Tea TUI is
planned and scaffolded under `rewrite/`.

Quick dev commands
- Build: make build
- Run tests: make test
- Coverage HTML: make coverage (writes coverage.html)
- Run CLI: go run . --help

Notes about current implementation
- Non-interactive streaming CLI: a simple, non-interactive CLI entrypoint exists at `cmd/gpt-cli` that supports `--stream` and smart TTY detection for ANSI rendering.
- Incremental rendering: `internal/ui` contains a markdown renderer that uses Glamour when running in a TTY and falls back to plain text for pipes.
- Providers: adapters include a plain HTTP adapter and an SDK-backed adapter; both support streaming and a non-streaming `Complete` path where applicable.
- Tests: unit and integration tests cover core buffering/streaming logic, providers, and rendering. An end-to-end test runs the built CLI against an httptest-based mock server.
- Mock server: a small `mock-openai` program under `mock-openai/` provides chunked and SSE variants for integration testing.
- CI: a GitHub Actions workflow is present at `.github/workflows/ci.yml` that runs unit and integration tests (starts the mock server in the integration job).

Current status
- Core: `internal/core` contains BufferManager and StreamReader to safely convert
  provider chunks into renderable fragments.
- Providers: `internal/core.BuildProviderOptions` reads env vars for provider config.
- Tests: unit tests for core logic exist and run with `go test ./internal/core`.
- TUI: Bubble Tea rewrite is in progress under `rewrite/`.

Contributing
- Run unit tests and keep changes small and well-tested.
- Add integration tests using the existing mock server for streaming scenarios.

How to run the CLI locally (examples)

Run a single prompt (non-interactive):
```
go run ./cmd/gpt-cli --provider http --base-url http://localhost:8081 "hello"
```

Run streaming output (progressive fragments):
```
go run ./cmd/gpt-cli --stream --provider http --base-url http://localhost:8081 "hello"
```

Run the mock server for local integration testing:
```
go run ./mock-openai -addr :8081
```

Notes
- Prefer the httptest-based mock servers in tests for CI; avoid running multiple mock server instances bound to the same port on the host.
- If you want the interactive TUI, use the `rewrite/` branch scaffold (planned work).

