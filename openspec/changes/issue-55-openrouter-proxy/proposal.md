## Why

SF-4 (varying the system prompt to test cache invalidation) is blocked because `opencode run` (headless mode) has no mechanism for dynamic system prompt injection at runtime. A transparent HTTP proxy between opencode and OpenRouter is the correct instrumentation layer: it intercepts requests before they reach the model, enabling controlled mutation without modifying opencode itself.

## What Changes

- New `spike/issue-45-opencode-cache/proxy/` directory containing a Go module that builds a single static binary
- Proxy implements `POST /v1/chat/completions` (OpenAI-compatible), forwarding to `https://openrouter.ai/api/v1/chat/completions` with the real API key
- SSE streaming response is forwarded verbatim via `io.Copy` — not buffered
- opencode authenticates to the proxy with a trivial static token (trusted loopback); the proxy holds the real `OPENROUTER_API_KEY` from env
- `spike/issue-45-opencode-cache/fixture/opencode.json` gains a custom provider entry pointing at `http://localhost:${PORT}`
- Proxy runs as a separate Docker container alongside the fixture container
- `spike/issue-45-opencode-cache/README.md` updated with two-container startup instructions
- Closes GitHub issue #55

## Capabilities

### New Capabilities

- `openrouter-proxy`: An OpenAI-compatible HTTP proxy that forwards chat completion requests to OpenRouter, handling SSE streaming and auth separation (opencode → proxy → OpenRouter)

### Modified Capabilities

_(none — no existing spec-level behavior changes)_

## Impact

- New Go toolchain dependency (build-time only; binary is statically linked)
- `spike/issue-45-opencode-cache/fixture/opencode.json` modified (custom provider added)
- No changes to opencode itself, the harness fixture logic, or the experiment methodology
- Prerequisite for SF-4b (#56), SF-4c (#57), SF-4d (#58)
