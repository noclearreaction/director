## Why

SF-4c (cache_control mutation) requires knowing the exact wire format of requests the proxy receives from opencode — specifically whether `messages[0].content` is a plain string or a content-parts array. Before mutating, we need to confirm what we are mutating.

## What Changes

- Add structured JSON logging to the proxy: each request/response pair written to a timestamped file in a mountable directory
- Log contains: full `messages` array, `model`, and response `usage` block (including `prompt_tokens_details.cached_tokens`)
- Log directory configurable via `LOG_DIR` env var (default `/logs`)
- No mutation of request or response — transparent passthrough preserved

## Capabilities

### New Capabilities

- `proxy-inspection`: Per-request JSON log files capturing inbound request body and outbound response usage fields, written to a mountable host directory

### Modified Capabilities

- `openrouter-proxy`: Request handling changes from streaming passthrough to buffer-then-forward to enable response logging; behavioral contract (transparent forwarding, SSE support) updated to reflect that SSE is now buffered

## Impact

- `spike/issue-45-opencode-cache/proxy/main.go` — core change: buffer request body, tee response, write log files
- `spike/issue-45-opencode-cache/proxy/AGENTS.md` — update env var and volume mount docs
- `spike/issue-45-opencode-cache/README.md` — update run commands to mount log volume
- No fixture changes required
