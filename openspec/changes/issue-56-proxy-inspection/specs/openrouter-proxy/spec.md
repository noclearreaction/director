## MODIFIED Requirements

### Requirement: Proxy forwards SSE streaming responses without buffering
The proxy SHALL buffer the full response body before forwarding to the client, to enable response logging. Streaming (SSE) responses are delivered as a single chunk rather than incrementally. This is acceptable because the fixture uses non-streaming mode (`stream: false`).

#### Scenario: Non-streaming response forwarded correctly
- **WHEN** OpenRouter returns a non-streaming JSON response
- **THEN** the full response body is buffered, logged, then forwarded to opencode with correct `Content-Type`

#### Scenario: SSE response delivered as single chunk
- **WHEN** OpenRouter returns a `text/event-stream` response
- **THEN** the full SSE body is accumulated and forwarded to opencode as a single write after the upstream response completes
