# GPT-CLI PR Review & Manual QA Findings

## Summary
This document records the results of a comprehensive code review and manual QA for the current Pull Request implementing shell mode, chat mode, enhanced configuration, and CLI improvements.

---

## Critical Issues 🚨

### 1. Shell Mode JSON Parsing Bug
- **Symptom:** Shell mode refinement fails for commands with ANSI escape codes (e.g., `\033[0;32mhello\033[0m`).
- **Root Cause:** Go's `json.Unmarshal` expects valid JSON, but the LLM sometimes returns single-quoted or improperly escaped strings, which are not valid JSON string escapes.
- **Impact:** Shell mode refinement fails for commands with ANSI escape codes or other non-JSON-compliant strings.
- **Action:** Harden the parser to handle escape sequences and single-quoted strings, or instruct the LLM to always use double quotes and valid JSON escapes.

### 2. Shell Mode Input Handling
- **Symptom:** After entering an invalid choice, the prompt repeats, but subsequent input is misparsed (e.g., entering "output" instead of "e/m/r/a").
- **Root Cause:** The CLI uses `fmt.Scanln(&choice)`, which can misparse multi-word or accidental input, leading to repeated invalid choice errors.
- **Impact:** User experience is degraded if the user mistypes or pastes unexpected input.
- **Action:** Improve input parsing to ignore invalid choices and prompt clearly for valid input.

---

## High Priority 📋

### 3. Chat Mode Exit Shortcut
- **Symptom:** Typing `q` exits the chat immediately, which is easy to do by accident.
- **Recommendation:** Remove `q` as an exit shortcut; keep only `Ctrl+C` or `Ctrl+D` for quitting.

---

## Medium Priority 💡

### 4. Shell Mode LLM Prompt
- **Action:** Update the system prompt to explicitly require valid JSON with double quotes and proper escaping for shell commands.

### 5. User Feedback
- **Action:** Add error messages or guidance when parsing fails, suggesting the user try a simpler command or edit manually.

---

## Deployment Readiness
- **Most features work as advertised.**
- **Shell mode refinement/parsing bug is a blocker for advanced shell suggestions.**
- **Chat mode is usable, but accidental exits are a UX risk.**

---

## Manual Testing Results

| Feature         | Status         | Notes/Issues                                      |
|-----------------|---------------|---------------------------------------------------|
| Shell Mode      | ✅/⚠️          | Works for simple commands; fails on ANSI escapes  |
| Chat Mode       | ✅/⚠️          | Works; accidental exit via `q` is risky           |
| Help Output     | ✅             | Clear and complete                                |
| CLI Flags       | ✅             | All modes available                               |
| Parsing/Input   | ⚠️             | Needs hardening for invalid/complex input         |

---

## Next Steps
1. Fix shell mode JSON parsing for escape sequences and single quotes.
2. Improve shell mode input handling for invalid choices.
3. Remove `q` as a chat exit shortcut.
4. (Optional) Update LLM prompt for stricter JSON compliance.

---

# Implementation Plan for PR Refinement

## 1. Shell Mode JSON Parsing Bug

Goal: Make shell mode robust against LLM responses containing ANSI escape codes, single quotes, or invalid JSON.

Steps:
- A. Harden the JSON parser in `internal/modes/shell.go`:
  - Improve `parseShellSuggestion` to:
    - Detect and safely convert single-quoted JSON to double-quoted JSON when appropriate.
    - Escape backslashes and other problematic characters before parsing.
    - Use a tolerant extraction approach: try extracting JSON with regex, then attempt fixes (unescape, replace single quotes) before failing.
    - Add a fallback: If parsing fails, display the raw command and explanation, and prompt the user to manually edit or execute.
  - Add unit tests in `internal/modes/shell_test.go` for edge cases (ANSI escapes, single quotes, malformed JSON).

- B. Update the LLM system prompt:
  - In `SuggestCommand`, clarify that the response must use double quotes and valid JSON escapes for shell commands.
  - Example addition to prompt: `IMPORTANT: Always use double quotes and valid JSON escapes for shell commands.`

## 2. Shell Mode Input Handling

Goal: Improve user experience when entering choices in shell mode.

Steps:
- A. Update input parsing in `promptUserAction` (in `internal/modes/shell.go`):
  - Replace `fmt.Scanln(&choice)` with `bufio.NewReader(os.Stdin).ReadString('\n')` for robust single-line input handling.
  - Trim whitespace and validate input strictly against allowed choices (`e`, `m`, `r`, `a`).
  - Normalize common full-word inputs (e.g., `execute` -> `e`, `edit` -> `m`, `refine` -> `r`, `abort` -> `a`).
  - On invalid input, display a clear error and re-prompt without misparsing.

- B. Add tests for invalid and multi-word input in `shell_test.go`.

## 3. Chat Mode Exit Shortcut

Goal: Prevent accidental exits from chat mode.

Steps:
- A. In `handleKeyMsg` in `internal/modes/chat.go`:
  - Remove `"q"` as a quit shortcut; keep only `Ctrl+C` and `Ctrl+D` for quitting.
  - Update help text in the chat UI to reflect the change.

- B. Add a test in `chat_test.go` to verify that `q` no longer exits the chat.

## 4. User Feedback for Shell Mode Parsing Errors

Goal: Guide users when shell suggestion parsing fails.

Steps:
- A. In `parseShellSuggestion` and `InteractiveMode` (in `internal/modes/shell.go`):
  - If parsing fails, display a user-friendly error:
    - "Could not parse the suggested command. Please try a simpler request or manually edit the command."
  - Offer the option to manually edit and execute the raw command.

- B. Add tests for error handling and user feedback in `shell_test.go`.

## General Steps

1. Update code in the relevant files (`internal/modes/shell.go`, `internal/modes/chat.go`, and their tests).
2. Add/modify unit tests to cover new edge cases and behaviors.
3. Update documentation and help text as needed (README, help output).
4. Manually test all modes after changes.
5. Commit changes and reference this plan in the PR refinement.

---

**This file should be referenced for the next PR refinement before merging.**
