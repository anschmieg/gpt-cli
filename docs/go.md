Go implementation notes

This document provides developer-facing notes for working on the Go rewrite.

Key packages
- internal/core
  - BufferManager: conservative buffering that flushes on newline and keeps
    fenced code blocks intact until closing fences are observed.
  - StreamReader: reads an `io.Reader` source and uses BufferManager to emit
    safe fragments.
  - BuildProviderOptions: reads env vars for OPENAI_/COPILOT_/GEMINI_ credentials.

How streaming works
1. Provider HTTP client reads the response body as a stream (chunked or SSE).
2. Chunks are fed into `StreamReader` (or directly into `BufferManager`).
3. `BufferManager` yields "safe fragments" suitable for incremental Markdown rendering.
4. The TUI or non-TTY renderer consumes fragments and renders them (Glamour is a good fit).

Developer workflow
- Run unit tests: `make test`
- Run package tests: `go test ./internal/core -v`
- Generate coverage: `make coverage` -> opens `coverage.html` in the workspace root.

Testing tips
- Use small chunk sizes in tests to validate boundary handling.
- Use `t.Setenv` (Go 1.17+) or helper functions to isolate env vars across tests.
- For integration tests, use `MOCK_SERVER_URL` and `GPT_CLI_TEST=1` to have tests talk to the repo mock server.

Next steps
- Implement provider HTTP streaming adapter that returns an `io.Reader` of chunks.
- Integrate Glamour for incremental rendering of fragments.
- Wire Bubble Tea TUI to subscribe to fragment channel and update view progressively.

