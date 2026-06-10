## Context

The proxy currently forwards requests transparently without capturing any record of what passes through. SF-4c requires mutating the `messages[0].content` field from a plain string to a content-parts array with `cache_control` markers. Before implementing mutation, we need to confirm the exact wire format opencode sends — specifically the shape of `messages[0].content` and that the response `usage` block includes `prompt_tokens_details.cached_tokens`.

The proxy is a single Go file using only stdlib. Logging must be added without breaking the existing transparent forwarding contract.

## Goals / Non-Goals

**Goals:**
- Capture the full inbound request body and outbound response body per request as JSON files
- Write files to a directory configurable via `LOG_DIR` env var, mountable from the host
- Confirm `messages[0].content` shape (plain string vs. content-parts array) from live opencode traffic
- Confirm `prompt_tokens_details.cached_tokens` field presence in response

**Non-Goals:**
- Request or response mutation (SF-4c)
- Log rotation, size limits, or retention policy — this is a spike tool
- SSE/streaming support — opencode's `build` agent uses non-streaming; buffering the full response is acceptable

## Decisions

**Buffer request body rather than tee**: Read the entire request body into memory with `io.ReadAll`, then forward using `bytes.NewReader`. Simple, no goroutine needed, acceptable for the small payloads in this spike.

**Buffer response body rather than tee**: Same approach — `io.ReadAll` on the upstream response, then log and write to the downstream client. This drops streaming/SSE passthrough for the logged path. Acceptable because opencode's `build` agent uses non-streaming (`stream: false`). If streaming were needed, a `io.TeeReader` with a goroutine would be required; not worth the complexity here.

**JSON files, one per request**: Named by UTC timestamp (`20060102T150405.000000000Z.json`). Flat directory. Simple to inspect with `cat` or `jq`. No database, no structured log aggregator.

**Non-JSON responses stored as string**: If the response body is not valid JSON (e.g., a partial SSE frame), it is stored as a JSON string value rather than failing. Keeps the log file valid JSON regardless.

**`LOG_DIR` env var, default `/logs`**: Consistent with Docker volume mount conventions. Directory is created on startup if absent.

## Risks / Trade-offs

**Memory usage for large responses**: Buffering the full response in memory before forwarding means large responses (e.g., long reasoning traces) are held in RAM. For this spike's payloads (~2KB responses), this is negligible. → No mitigation needed at spike scale.

**Log files accumulate indefinitely**: No rotation or cleanup. → Acceptable for a spike; user mounts a host directory and clears manually.

**SSE passthrough broken**: The previous `io.Copy` + `http.Flusher` loop is replaced by full buffering. Streaming responses will be delivered in one chunk rather than incrementally. → Not an issue because the fixture uses non-streaming mode.
