# Requirements for gpt-cli

Plan: produce a single, comprehensive requirements spec covering functional behavior, inputs/outputs, edge cases, non-functional constraints, test/CI needs, and implementation notes for a Bubble Tea–based app that supports non-interactive, interactive, semi‑interactive suggest, and stdin/stdout modes.

## Checklist (visible requirements)
- Modes: non-interactive, interactive chat (TUI), semi-interactive suggestion (TUI), stdin/stdout support
- Streaming: incremental provider streaming support + realtime markdown rendering
- Partial-token buffering: handle markdown tokens crossing chunks safely
- Suggestion workflow: present suggestion, allow Execute / Edit & Execute / Refine / Abort
- Provider adapters: OpenAI-compatible, Copilot, Gemini, mock server support
- Config: env vars, file defaults, CLI flags, interactive config UI
- Tests: unit + integration using existing mock server
- CI: build, unit test, integration (mock) jobs
- UX: TTY detection, non-TTY fallback (plain streaming/pipe-friendly)
- Security & safety: sandbox suggestion execution, safe default outputs, explicit user confirmation
- Performance: low-latency, backpressure-aware streaming, modest binary size

1) High-level functional requirements
- Single binary using Bubble Tea for all interactive behavior. Must also support non-interactive runs and stdin/stdout usage.
- Modes:
  - Non-interactive (default): run a single prompt, print rendered output to stdout (ANSI if TTY, plain text if piped). Support `--stream` to show progressive output when provider streams.
  - Interactive chat (TUI): multi-turn chat UI with prompt input, message history, streaming rendering, model selection, system prompt editing, and settings panel.
  - Semi-interactive suggest (TUI): request suggestions from the model, parse suggestion JSON contract, show candidate command(s) and present UI choices: Execute, Edit (inline text input), Refine (send new prompt), Abort. Executing must require explicit confirm.
  - stdin mode: read prompt from stdin when no prompt arg or when `-` used, and behave non-interactively (support streaming).
- Provider plumbing:
  - Adapters for OpenAI-compatible, GitHub Copilot, Google Gemini.
  - Use `MOCK_SERVER_URL` and `GPT_CLI_TEST` for integration tests.
  - Build provider options from env vars (OPENAI_API_KEY, COPILOT_API_KEY, GEMINI_API_KEY, bases).
- Output & rendering:
  - Markdown rendering to ANSI in TTY. Use incremental rendering for streaming. If stdout is not TTY, output plain text or JSON as appropriate.
  - For suggestion mode, emit a JSON suggestion object when requested (scriptable mode), and present the richer UI when interactive.
- CLI flags / UI entry points:
  - Flags for provider, model, temperature, system, file upload, stream, suggest, interactive, verbose, markdown on/off, retry-model, help.
  - `tui` or `interactive` subcommand to start full TUI explicitly.
- Error handling:
  - Clear user-facing errors for network/provider issues.
  - Retry logic for unsupported model when `--retry-model` enabled.
  - For suggestion parse failures show safe fallback suggestion message (structured).

2) Input / output contract (tiny "contracts")
- Inputs:
  - CLI args and flags
  - stdin (prompt text)
  - env vars for keys and endpoints
  - interactive keyboard events via Bubble Tea
- Outputs:
  - Streamed ANSI-rendered markdown to TTY
  - Plain text when piped/non-TTY
  - Structured JSON for suggestion-mode machine output (when asked)
  - Exit codes: 0 on success, non-zero on failures (1 for generic errors, specific codes optional)
- Data shapes:
  - Suggestion JSON: { suggestions: [{ command, description, category, risk, args[] }], context: string, safe: bool }

3) Edge cases & streaming/partial-token policy
- Problem: markdown tokens may split across provider chunks.
- Buffering strategy (requirements):
  - Maintain a small buffer and only pass "safe fragments" to renderer.
  - Flush boundaries:
    - Always safe to flush on blank-line boundaries or newline-terminated paragraphs.
    - For fenced code blocks: detect opening fence ("```" or "~~~"); do not render until matching closing fence is received. While inside fence show a "loading" placeholder or append unrendered block content in a monospace/plain area.
    - For inline tokens (bold/italic, backticks): prefer line-based flushes; if a chunk ends in a potential mid-token (unclosed `*`, `_`, or backtick), hold until either token closed or a safe timeout/size threshold reached.
  - Timeouts and max-buffer:
    - If buffer grows beyond N KB (configurable), flush anyway with best-effort and mark as partial.
    - Provide a graceful degrade: if forced flush occurs within a token, render as-is; the TUI will later correct when closing token arrives.
- Streaming model:
  - Consumer goroutine reads from provider (SSE/chunked/streaming JSON). Producer sends raw chunks into buffer manager which yields safe fragments to renderer or TUI model messages.

4) Suggestion mode UX & flows
- Accept provider suggestion JSON or markdown that includes suggestion content.
- Parse suggestions robustly; detect parsing errors and fall back to a single safe suggestion:
  - { command: "echo 'Unable to generate safe suggestions'", ... }
- Interactive choices:
  - Preview suggestion (highlight command)
  - Select option:
    - Execute: (a) show confirmation modal; (b) if confirmed, execute in a sandboxed manner (do not execute destructive defaults); (c) show execution output in TUI.
    - Edit & Execute: open inline editable textinput prefilled with the suggestion; on confirm execute.
    - Refine: open prompt input to ask LLM for a refined suggestion (loop back into suggestion generator).
    - Abort: discard suggestion and return to prompt.
- Safety:
  - By default do not auto-execute untrusted suggestions. Always require explicit confirm.
  - Consider a "trusted" flag or allow command whitelisting in config for advanced users.

5) Non-interactive & TTY detection
- Detect `isatty` on stdin/stdout.
- If piped (non-TTY): do not start Bubble Tea UI; instead use non-interactive streaming renderer and JSON output options (`--suggest --json`).
- Allow explicit override flags: `--no-tui`, `--tui`.

6) Tests and verification
- Unit tests:
  - Buffering & safe-fragment logic (cases: partial fence, inline token breaks, long streams).
  - Suggestion JSON parsing & fallback.
  - Provider option building from env vars (hermetic tests).
  - Markdown renderer render correctness for typical tokens.
- Integration tests using existing mock server:
  - Streaming integration: mock server sends chunk sequences that include partial tokens; assert TUI/model behavior or output correctness.
  - Suggestion flow: mock server returns suggestion payload; verify TUI actions behave or simulate user input.
- CI:
  - Unit job (go test ./...).
  - Integration job run on PR/main that starts mock server and runs integration tests (limit network).
  - Coverage artifact job optional.

7) Non-functional requirements & constraints
- Binary size:
  - Single Bubble Tea binary will be larger than a minimal CLI, but acceptable. Use `-ldflags "-s -w"` to strip debug.
  - If size becomes critical we can later split into `gpt-cli` + `gpt-cli-tui`.
- Performance:
  - Streaming latency < 200ms perceived for small chunks; network dominates.
  - Backpressure: rendering must not block reading; use buffered channels and bounded worker goroutines.
- Resource limits:
  - Memory: bound buffer sizes, limit concurrent streams.
- Security:
  - Never execute suggestions without explicit confirm.
  - Avoid executing shell commands in the same process environment without explicit user opt-in.

8) Libraries & implementation notes (suggested)
- TUI: Bubble Tea + Bubbles + Lipgloss
- Markdown rendering: Glamour (charmbracelet/glamour) for Markdown→ANSI
- SSE / streaming parsing: custom SSE reader or use simple chunk reader for chunked responses; for JSON streaming use incremental decoder (json.Decoder with Token).
- Tests: Go's testing package with `t.Setenv` and the existing `startMockTestServer()` helper
- Build flags: use strip flags and optionally `upx` for distribution size

9) Acceptance criteria (concrete)
- All unit tests pass locally and on CI.
- Non-interactive mode:
  - `gpt-cli "small prompt"` prints rendered output and returns 0.
  - `echo "prompt" | gpt-cli` reads stdin and prints plain text (non-ANSI) if piped.
- Streaming:
  - `gpt-cli --stream` shows progressive markdown rendering with no mid-token corruption on standard chunk cases from mock.
- Interactive chat:
  - Starting TUI (`gpt-cli tui` or `gpt-cli --interactive`) opens a responsive chat with streaming messages, and new message input works.
- Suggest mode:
  - `gpt-cli suggest` (TUI) shows suggestion, user can Execute / Edit / Refine / Abort. Execution requires explicit confirmation.
  - `gpt-cli --suggest --json` outputs suggestion JSON machine-readable (non-TTY).
- Robustness:
  - When provider returns malformed JSON for suggestions, app shows safe fallback suggestion and does not crash.

10) Assumptions & decisions to confirm
- Use Bubble Tea as the single UI framework and implement non-TTY fallback mode in the same binary (ok per earlier discussion unless you want split binary).
- Use Glamour for markdown rendering.
- Use the existing mock server and tests as the integration harness.
- Prefer `go test ./... -v` + CI GitHub Actions updated for this repo layout.
