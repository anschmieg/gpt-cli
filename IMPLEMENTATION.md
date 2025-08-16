– und unterstützt konfigurierbare Parameter wie Modellwahl, Temperatur,
# Implementation checklist

This file is a compact checklist covering project structure, goals, required features, current status, and small actionable tasks you can pick and run immediately.

## Project overview
- [x] CLI parsing (`cli.ts`) — parses flags, builds `CoreConfig` and calls `runCore`.
- [x] Core orchestration (`core.ts`) — builds provider call, handles responses, prints output.
- [x] Providers (`providers/*.ts`) — adapter modules for OpenAI-compatible APIs (openai, copilot, gemini).
- [x] Mock server (`mock-openai/mock-server.ts`) — local mock OpenAI-compatible server used for integration tests.
- [~] Utils (`utils/*.ts`) — logging, markdown rendering, small helpers. (partial)

## Goals
- [ ] Reliable non-streaming responses (text/markdown).
- [ ] Streaming support (SSE / incremental parsing).
- [ ] Robust markdown rendering (non-streaming and streaming).
- [ ] Shell suggestion mode (safe suggestion output, no execution by default).
- [x] CI coverage generation and artifacts.

## Feature checklist & status

1) Streaming + non-streaming responses
- [x] Mock server provides SSE for streaming.
- [ ] Provider adapters expose streaming path (AsyncGenerator or callback).
- [ ] Core supports streaming provider output and progressive printing.
- Next small task: implement streaming read in `providers/api_openai_compatible.ts` and add a unit/integration test.

2) Markdown parsing (non-streaming)
- [ ] Implement `renderMarkdown` in `utils/markdown.ts` to produce ANSI-decorated output for terminals.
- Next small task: add simple renderer (headings, emphasis, lists, code fences) and tests.

3) Markdown parsing for streaming responses
- [ ] Implement streaming tokenizer/renderer that can handle token boundaries across chunks.
- Next small task: implement buffered chunk renderer (buffer N KB, parse, render) as interim solution.

4) Shell suggestion mode
- [ ] Add `--suggest` or `--mode=suggest` flag in `cli.ts` and wire through `core.ts`.
- [ ] Implement a deterministic JSON response contract from provider prompts for suggestions.
- Next small task: add CLI flag and parse-only path that returns parsed suggestion JSON (no execution).

5) Markdown visual rendering (streaming)
- [ ] Integrate streaming renderer into CLI printing for progressive visual output.
- Next small task: create a streaming test harness using the mock server and verify output parity with non-streamed renderer.

## Tests to add
- [ ] Unit tests for `renderMarkdown` (non-streaming snapshots).
- [ ] Integration test: mock server streaming -> streaming renderer -> final output matches non-streaming.
- [ ] Shell suggestion tests: prompt -> parsed JSON suggestion.

## CI checklist
- [x] Separate unit and integration jobs in `.github/workflows/ci.yml`.
- [x] Coverage job writes LCOV to `coverage/lcov.info` and uploads artifact.
- [x] Ensure integration job is gated for tags/releases (it currently is).
- Next small task: add optional on-demand workflow dispatch to run integration+coverage for feature branches.

## Who/when
- Pick one of the "next small task" items above and I will implement it, add tests, and run the suite locally.

