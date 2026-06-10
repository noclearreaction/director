## Context

opencode communicates with AI providers using the OpenAI-compatible chat completions API (`POST /v1/chat/completions`). Its current upstream is OrcaRouter, which itself proxies to OpenRouter. Because opencode is a black box from the experiment's perspective, we cannot intercept or mutate what it sends at the application layer. The proxy sits transparently in the network path, giving us full control over request content before it reaches the model.

The proxy will be used across the SF-4 sub-spikes: baseline passthrough (#55), traffic inspection (#56), mutation design (#57), and Go template mutation (#58). This design covers only the baseline passthrough (#55).

## Goals / Non-Goals

**Goals:**
- Transparent passthrough of all `POST /v1/chat/completions` requests to OpenRouter
- SSE streaming response forwarded without buffering
- Auth separation: opencode uses a trivial token to reach the proxy; proxy uses the real `OPENROUTER_API_KEY`
- Zero functional change to cache behavior vs. direct path (SF-3 baseline must hold)
- Single static Go binary, Dockerfile for container deployment

**Non-Goals:**
- Request/response logging (SF-4b, #56)
- Prompt mutation (SF-4c/d, #57–#58)
- Any other HTTP methods or endpoints beyond `POST /v1/chat/completions`
- TLS termination (loopback HTTP only)
- Production hardening (rate limiting, auth enforcement, etc.)

## Decisions

### Language: Go
Go was chosen over TypeScript/Deno for this proxy. [`net/http`](https://pkg.go.dev/net/http) and [`io.Copy`](https://pkg.go.dev/io#Copy) handle SSE streaming natively with no buffering risk. [`text/template`](https://pkg.go.dev/text/template) (needed in SF-4d) is stdlib. Output is a single static binary with no runtime dependency. Deno/TypeScript were considered; Go wins on proxy-specific merits (streaming, single binary, stdlib completeness).

Install Go on Ubuntu/WSL: [https://go.dev/doc/install](https://go.dev/doc/install) (download the Linux tarball — the apt package may be too old).

### SSE forwarding: `io.Copy` not buffered response
The proxy must not buffer the full response before forwarding — that would break streaming and add latency. [`io.Copy`](https://pkg.go.dev/io#Copy) from the upstream `http.Response.Body` directly to the `ResponseWriter` is the correct pattern. The `Content-Type: text/event-stream` and `Transfer-Encoding` headers must be forwarded verbatim. [`http.Flusher`](https://pkg.go.dev/net/http#Flusher) is called after each write to ensure chunks reach opencode immediately.

Reference: [MDN — Server-Sent Events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events)

### Auth model: static env token for proxy, real key for upstream
opencode's custom provider config accepts a bearer token. The proxy validates an `PROXY_TOKEN` env var (optional; if unset, accepts all requests — appropriate for loopback-only deployment). The proxy replaces the `Authorization` header with `Bearer ${OPENROUTER_API_KEY}` before forwarding. This keeps the real key out of opencode's config entirely.

opencode custom provider config reference: [https://opencode.ai/docs/providers](https://opencode.ai/docs/providers)
OpenRouter auth docs: [https://openrouter.ai/docs/api-reference/authentication](https://openrouter.ai/docs/api-reference/authentication)

### Deployment: separate container, shared Docker network
The proxy and the fixture container run on the same Docker user-defined network. opencode's custom provider URL uses the proxy's container name as hostname (e.g., `http://openrouter-proxy:8080`). This avoids host IP lookup complexity and works identically on any host.

Reference: [Docker user-defined bridge networks](https://docs.docker.com/network/drivers/bridge/)

### Model selection: forwarded from opencode request
The proxy forwards the `model` field from the request body verbatim. It does not override or validate the model. This allows the fixture's `opencode.json` to control model selection as before.

OpenRouter model IDs: [https://openrouter.ai/models](https://openrouter.ai/models)
OpenAI chat completions wire format (reference): [https://platform.openai.com/docs/api-reference/chat/create](https://platform.openai.com/docs/api-reference/chat/create)

## Risks / Trade-offs

- **SSE flush not triggered** → [`http.ResponseWriter`](https://pkg.go.dev/net/http#ResponseWriter) in Go does not auto-flush; explicit [`Flusher.Flush()`](https://pkg.go.dev/net/http#Flusher) after each write is required. If omitted, opencode will hang waiting for the full response.
- **Header conflicts** → `Content-Length` must be stripped when forwarding chunked SSE; the proxy must not forward it or the client will truncate the stream. Use [`http.Header.Del`](https://pkg.go.dev/net/http#Header.Del)`("Content-Length")` before writing.
- **OpenRouter extra headers** → OpenRouter returns non-standard headers (`x-ratelimit-*`, etc.) which must be forwarded to avoid opencode SDK parse errors on unexpected missing fields. Forward all response headers verbatim.
- **Go not installed in WSL** → Tasks include Go installation steps for Ubuntu/WSL. Go 1.22+ required. See [https://go.dev/doc/install](https://go.dev/doc/install).

## Open Questions

_(none — design is sufficient to begin implementation)_
