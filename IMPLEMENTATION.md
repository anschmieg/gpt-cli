# Implementation Checklist

This file provides a clear and actionable checklist to guide development and
maintenance, organized by project structure, goals, features, tests, and CI. All
action items are tracked with checkboxes for quick reference and status.

## Project Overview

- [x] CLI parsing (`cli.ts`) — parses flags, builds `CoreConfig`, calls
      `runCore`.
- [x] Core orchestration (`core.ts`) — builds provider call, handles responses,
      prints output; returns structured results (CLI handles exit).
- [x] Adapters (`adapters/*.ts`) — runtime adapter entrypoints for
      OpenAI-compatible APIs.
- [x] Mock server (`tests/mock-server-openai/mock-server.ts`) — local mock
      OpenAI-compatible server for integration tests.
- [x] Utils (`utils/*.ts`) — logging, markdown rendering, helpers. _(Partial;
  logging is now permission-safe.)_

## Project Goals

- [x] Reliable non-streaming responses (text/markdown).
- [~] Streaming support (SSE/incremental parsing). _(Basic functionality;
  improved rendering pending.)_
- [ ] Robust markdown rendering (for streamed and non-streamed content).
- [ ] Shell suggestion mode (safe output, no execution by default).
- [x] CI coverage generation and artifacts.

## Feature Checklist & Status

**Streaming & Non-streaming Responses**

- [x] Mock server provides SSE for streaming.
- [x] Adapters expose streaming (AsyncGenerator or callback).
- [x] Core supports streaming output and progressive printing.
- [x] Streaming read in `src/providers/api_openai_compatible.ts` with tests
      (integration gated by Deno permissions).

**Shell Suggestion Mode**

- [ ] Add `--suggest` or `--mode=suggest` flag in `cli.ts` and wire through
      `core.ts`.
- [ ] Deterministic JSON contract from provider for suggestions.
  - [ ] CLI flag and parse-only path, returning parsed JSON (no execution).

**Markdown Parsing (Non-streaming)**

- [ ] Implement `renderMarkdown` in `utils/markdown.ts` for ANSI-decorated
      terminal output.
  - [ ] Renderer for headings, emphasis, lists, code fences.
  - [ ] Add unit tests for renderer.

**Markdown Parsing (Streaming)**

- [ ] Streaming tokenizer/renderer that handles token boundaries across chunks.
  - [ ] Buffered chunk renderer (buffer N KB, parse, render) as interim.

**Markdown Visual Rendering (Streaming)**

- [ ] Integrate streaming renderer into CLI for progressive output.
  - [ ] Create streaming test harness using mock server.
  - [ ] Verify output parity with non-streamed renderer.

## Tests

- [x] Unit tests for adapter-utils (`normalizeProviderError`).
- [x] Adapter-shape runtime test (validates exports).
- [x] Unit test for auto-retry when model is not supported.
- [ ] Add tests for core and utils.
- [ ] Unit tests for `renderMarkdown` (non-streaming snapshots).
- [ ] Integration test: mock server streaming → streaming renderer → parity
      check.
- [ ] Shell suggestion mode tests: prompt → parsed JSON.

## CI Checklist

- [x] Separate unit/integration jobs in `.github/workflows/ci.yml`.
- [x] Coverage job writes LCOV to `coverage/lcov.info` and uploads artifact.
- [x] Integration job gated for tags/releases.
- [ ] Optional on-demand workflow dispatch for integration+coverage on feature
      branches.

## Prioritized TODOs

**P0 (High - Correctness/Test Stability)**

- [x] Centralize shared types in `src/providers/types.ts`.
- [x] Remove library-level `Deno.exit` calls; only CLI handles exit.
- [x] Deduplicate provider modules under `src/providers/`.

**P1 (Important - Core Features & DX)**

- [x] Formalize provider adapter contract: TS interface & runtime checks.
- [ ] Add `StreamRenderer` abstraction and wire into `runCore`.
- [ ] Ensure streaming writes respect stdout backpressure
      (`await Deno.stdout.write(...)`).

**P2 (Tests/CI Hygiene)**

- [ ] Integration test harness docs and `scripts/test-integration.sh` showing
      required Deno flags.
- [x] Harden mock server helpers.
- [ ] Test for generator cancellation (simulate SIGINT).

**P3 (UX/Features)**

- [ ] Implement streaming markdown renderer (handles token
      boundaries/incremental rendering).
- [ ] Wire CLI flag (`--stream`) and suggestion mode (`--mode=suggest`).
- [ ] Add snapshot tests for rendered markdown (streamed & non-streamed).

## Recent Improvements

- [x] Centralized provider types and adapter interface.
- [x] Centralized error handling (`normalizeProviderError`, etc).
- [x] Providers refactored to use `ProviderOptions`.
- [x] Runtime validation of adapters in tests.
- [x] CLI `--auto-retry-model` for automatic model retry.
- [x] Logging is now permission-safe.

## Security & Permissions

- [x] Process exit calls only in CLI; library code should not exit.
- [x] Tests avoid permissions where possible; prefer dependency injection and
      in-process helpers.

## Efficiency & Performance

      (`await Deno.stdout.write(...)`) recommended for larger streams.

## Maintainability Recommendations


## Quick Wins (Recently Completed)


## Quick Wins (Available To Pick)

      `runCore`.

## Implementation Review (automated)

Summary: focused review of structure, methodologies, and code efficiency.

- [x] Structure: logical and effective
      - Notes: Clear layering (CLI → core → adapters) and centralized types.
      - Action: add CONTRIBUTING.md describing where to add adapters (optional).

- [~] Methodology: appropriate but incomplete
      - Notes: Adapter pattern and error normalization are sound. Streaming
            renderer abstraction remains missing.
      - Action: implement `StreamRenderer` interface and wire into `runCore`.

- [~] Efficiency: mostly good, minor opportunities
      - Notes: Streaming/backpressure handled in places; audit required for
            larger payloads. Some duplicated helpers may exist across `src/` and
            `tests/`.
      - Action: add `scripts/audit-duplicates.ts` to flag consolidation targets.

- [ ] Redundancy/leftovers: cleanup recommended
      - Notes: Pre-commit hook iterated across implementations (bash/python/deno);
            ensure only final hook remains and remove legacy artifacts.
      - Action: remove obsolete hook files and document the Deno hook in README.

Next steps (3 small PRs):

- [ ] Add `StreamRenderer` interface + no-op implementation and wire into `runCore`.
- [ ] Add `scripts/audit-duplicates.ts` and run to consolidate helpers.
- [ ] Remove legacy hook scripts and add README section on the Deno pre-commit hook.
## Requirements Coverage

- [x] Centralized shared types
- [x] Remove library-level process exits
- [x] Deduplicate provider modules
- [x] Provider adapter contract (TS + runtime)
- [ ] StreamRenderer
- [ ] Streaming markdown renderer
