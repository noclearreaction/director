## Context

SF-4b wrote per-request JSON files. The correct granularity is per-session: all turns of one conversation in one file, appended as they arrive. This makes a session readable as a linear transcript.

The challenge is that the proxy has no explicit session ID from opencode — opencode sends no `X-Session-ID` header. Session grouping must be inferred from request content.

## Goals / Non-Goals

**Goals:**
- One NDJSON file per session (one JSON object per line, appended per request)
- Requests from the same opencode session land in the same file
- File is readable as a linear transcript: turn 1 on line 1, turn 2 on line 2, etc.

**Non-Goals:**
- Perfect session isolation — close enough for inspection purposes
- Handling concurrent sessions (spike runs one session at a time)
- Log rotation or size limits

## Decisions

**Session key = SHA-256 prefix of `messages[0].content`**: The system prompt is stable across all turns of a session and unique between the agent request and title-gen request. Hashing the first 512 bytes gives a compact, collision-resistant key. Title-gen requests (short, different system prompt) get a different key and land in a separate file.

**NDJSON (newline-delimited JSON)**: One JSON object per line, appended. Trivially parseable with `while read line; do echo $line | jq .; done` or Python. Avoids the complexity of maintaining a valid JSON array across concurrent writes.

**Turn number from `messages` length**: `(len(messages) + 1) / 2` gives the human turn number. Turn 1 has 2 messages (system + user), turn 2 has 4, etc.

**Mutex on file append**: Multiple concurrent requests (title-gen + agent) write to different files by design (different system prompts → different keys), so no mutex is needed between them. A per-file mutex guards against the unlikely case of two requests with identical system prompts racing.

## Risks / Trade-offs

**Title-gen still produces a separate file**: The title-generation request has a different system prompt, so it gets its own session file. This is acceptable — it separates concerns cleanly. The agent session file contains only agent turns.

**Key derived from content, not a real session ID**: If two different opencode sessions happen to use identical system prompts, they'll share a file. At spike scale (one session at a time), this cannot happen.
