## ADDED Requirements

### Requirement: Proxy accepts OpenAI-compatible chat completion requests
The proxy SHALL accept `POST /v1/chat/completions` requests using the OpenAI chat completions wire format. All other routes SHALL return HTTP 404.

#### Scenario: Valid chat completion request forwarded
- **WHEN** opencode sends `POST /v1/chat/completions` with a valid JSON body
- **THEN** the proxy forwards the request to OpenRouter and returns the response to opencode

#### Scenario: Unknown route rejected
- **WHEN** a request arrives at any path other than `/v1/chat/completions`
- **THEN** the proxy returns HTTP 404

---

### Requirement: Proxy forwards SSE streaming responses without buffering
The proxy SHALL forward Server-Sent Events streaming responses from OpenRouter to the client incrementally, without accumulating the full response body before forwarding.

#### Scenario: Streaming response forwarded in real time
- **WHEN** OpenRouter returns a `text/event-stream` response
- **THEN** each SSE chunk is forwarded to opencode as it arrives, with no buffering delay

#### Scenario: Non-streaming response forwarded correctly
- **WHEN** OpenRouter returns a non-streaming JSON response
- **THEN** the full response body is forwarded to opencode with correct `Content-Type`

---

### Requirement: Proxy replaces Authorization header with OpenRouter API key
The proxy SHALL strip the `Authorization` header from incoming requests and replace it with `Bearer ${OPENROUTER_API_KEY}` before forwarding to OpenRouter. The real API key SHALL only be present in the proxy's environment, never in opencode's configuration.

#### Scenario: Downstream key replaced with upstream key
- **WHEN** opencode sends a request with any `Authorization` header value
- **THEN** the proxy forwards the request to OpenRouter with `Authorization: Bearer ${OPENROUTER_API_KEY}`

#### Scenario: Missing OPENROUTER_API_KEY prevents startup
- **WHEN** the proxy starts and `OPENROUTER_API_KEY` is not set in the environment
- **THEN** the proxy exits with a non-zero exit code and a descriptive error message

---

### Requirement: Proxy and fixture run on a shared Docker network
The proxy SHALL be deployable as a Docker container on a user-defined Docker network, reachable by the fixture container using the proxy's container name as hostname.

#### Scenario: opencode connects to proxy by container hostname
- **WHEN** the fixture container runs with `OPENROUTER_PROXY_URL=http://openrouter-proxy:8080`
- **THEN** opencode successfully routes all AI requests through the proxy

#### Scenario: Proxy and fixture started with documented two-container workflow
- **WHEN** a user follows the README startup instructions
- **THEN** both containers start, connect, and a single `opencode run` turn completes successfully through the proxy
