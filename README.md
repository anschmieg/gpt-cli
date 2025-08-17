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

Current status
- Core: `internal/core` contains BufferManager and StreamReader to safely convert
  provider chunks into renderable fragments.
- Providers: `internal/core.BuildProviderOptions` reads env vars for provider config.
- Tests: unit tests for core logic exist and run with `go test ./internal/core`.
- TUI: Bubble Tea rewrite is in progress under `rewrite/`.

Contributing
- Run unit tests and keep changes small and well-tested.
- Add integration tests using the existing mock server for streaming scenarios.

