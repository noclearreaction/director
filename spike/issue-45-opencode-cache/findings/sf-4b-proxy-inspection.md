# SF-4b Proxy Traffic Inspection Findings

opencode 1.16.2 | model: `google/gemini-2.5-flash` via `openrouter-proxy` | date: 2026-06-10

---

## Method

The proxy was updated to buffer and log each request/response pair as a JSON file. A 3-turn session was run through the proxy with the log directory mounted to the host. Log files were inspected with Python.

---

## Finding 1: opencode always uses streaming

**All** requests from opencode arrive with `"stream": true`, including the `build` agent. There is no non-streaming path in normal opencode operation. The proxy's buffered-response approach means SSE responses are accumulated and delivered as a single chunk — this has no visible effect on opencode since it parses the stream on receipt.

---

## Finding 2: `messages[0].content` is a plain string

The system message content is sent as a single plain string, not a content-parts array:

```json
{
  "role": "system",
  "content": "# Experiment Agent\n\nYou are a minimal experiment agent..."
}
```

**Implication for SF-4c**: To inject `cache_control`, the proxy must transform this:

```json
{ "role": "system", "content": "<plain string>" }
```

into this:

```json
{
  "role": "system",
  "content": [
    { "type": "text", "text": "<plain string>", "cache_control": { "type": "ephemeral" } }
  ]
}
```

This is the exact mutation the proxy must perform to activate Gemini explicit caching through OpenRouter.

---

## Finding 3: opencode sends two concurrent requests per turn

Each `opencode run` turn produces **two** simultaneous requests:

1. A title-generation request (short system prompt, `stream: true`)
2. The actual agent request (full system prompt, `stream: true`)

Both use `google/gemini-2.5-flash`. The title request uses a different system prompt ("You are a title generator...") and must not have `cache_control` injected, or the mutation must be conditional on system prompt content.

**Implication for SF-4c**: The proxy should only inject `cache_control` when the system prompt contains the experiment fixture content, or conditionally based on a header or env var.

---

## Finding 4: Session history grows as plain strings

Across turns, the `messages` array grows with alternating `user`/`assistant` entries, all with plain string `content`. The system message (`messages[0]`) remains constant across turns.

Turn 1: 2 messages (system + user)  
Turn 2: 4 messages (system + user + assistant + user)  
Turn 3: 6 messages (system + user + assistant + user + assistant + user)  

**Implication for SF-4c**: Injecting `cache_control` only on `messages[0]` is sufficient and stable — it never changes between turns.

---

## Finding 5: Usage fields present but cached_tokens=0 without cache_control

The SSE response stream contains a `[DONE]` frame and usage data is not surfaced in the streamed chunks in a way the current logger captures (the full SSE body is stored as a string). To read `cached_tokens` from streaming responses, the proxy would need to parse SSE frames. However, the `opencode db` query on the `message` table shows `cache.read` which opencode extracts from the stream internally.

Direct non-streaming tests (see SF-4b precursor work) confirmed:
- Without `cache_control`: `cached_tokens = 0` even at 1166 prompt tokens
- With `cache_control` on system prompt block: `cached_tokens = 1164`, cost reduced 70%

---

## Summary

| Question | Answer |
|---|---|
| `messages[0].content` shape | Plain string |
| opencode streaming mode | Always `stream: true` |
| Requests per turn | 2 (title + agent) |
| Cache without `cache_control` | 0 |
| Cache with `cache_control` | 1164 tokens, 70% cost reduction |
| SF-4c mutation target | `messages[0].content`: string → content-parts array with `cache_control` |
| Mutation scope guard | Inject only when system prompt matches fixture identity (not title generator) |
