## MODIFIED Requirements

### Requirement: Per-request JSON log files
The proxy SHALL append one JSON line per forwarded request to a session-keyed NDJSON file in the configured log directory. Each line SHALL contain the timestamp, turn number, full inbound request body, and full outbound response body (or error). Multiple requests belonging to the same session SHALL be appended to the same file in order.

#### Scenario: Requests from the same session land in the same file
- **WHEN** three consecutive turns of one opencode session pass through the proxy
- **THEN** all three are appended as separate lines to a single `<session-key>.ndjson` file

#### Scenario: Failed upstream request appended to session file
- **WHEN** the upstream request fails
- **THEN** a JSON line is appended containing `timestamp`, `turn`, `request`, and `error` fields; `response` is omitted

#### Scenario: Non-JSON response stored as string
- **WHEN** the upstream response body is not valid JSON (e.g. SSE stream)
- **THEN** the `response` field SHALL contain the raw body as a JSON string value

### Requirement: Session key derived from system prompt
The proxy SHALL derive the session key by taking the SHA-256 hash of the first 512 bytes of `messages[0].content`. Requests with different system prompts (e.g. title-generation vs. agent) SHALL produce different session keys and be written to different files.

#### Scenario: Agent request and title-gen request produce separate files
- **WHEN** opencode sends a title-generation request and an agent request in the same turn
- **THEN** each lands in a separate NDJSON file based on their respective system prompt hashes

### Requirement: Log file naming
Log files SHALL be named `<first-8-chars-of-session-key>.ndjson` in the configured log directory.

#### Scenario: File name is stable across turns
- **WHEN** turn 2 arrives with the same system prompt as turn 1
- **THEN** it is appended to the same file (same 8-char prefix, same filename)
