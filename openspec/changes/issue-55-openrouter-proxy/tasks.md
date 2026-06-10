## 1. Environment Setup

- [ ] 1.1 Install Go 1.22+ on Ubuntu/WSL: `sudo apt update && sudo apt install -y golang-go` (verify: `go version` should show 1.22+; if apt version is older, install from https://go.dev/dl/ manually)
- [ ] 1.2 Verify Docker and `docker network` commands are available: `docker --version && docker network ls`
- [ ] 1.3 Create Docker user-defined network for spike containers: `docker network create spike-net`

## 2. Proxy Scaffold

- [ ] 2.1 Create `spike/issue-45-opencode-cache/proxy/` directory
- [ ] 2.2 Initialize Go module: `go mod init github.com/noclearreaction/symphony-director/spike/proxy` inside `proxy/`
- [ ] 2.3 Create `proxy/main.go` with HTTP server skeleton: `net/http` listener on `PORT` env var (default `8080`), single route `POST /v1/chat/completions`, all other routes return 404

## 3. Core Proxy Logic

- [ ] 3.1 Implement startup validation: exit with error if `OPENROUTER_API_KEY` env var is not set
- [ ] 3.2 Implement request forwarding: copy incoming request body, set `Authorization: Bearer ${OPENROUTER_API_KEY}`, forward to `https://openrouter.ai/api/v1/chat/completions`
- [ ] 3.3 Implement SSE streaming passthrough: forward all response headers verbatim (strip `Content-Length`), copy response body using `io.Copy` with `http.Flusher.Flush()` after each write
- [ ] 3.4 Verify non-streaming response path also works (full JSON body forwarded correctly)

## 4. Docker Container

- [ ] 4.1 Create `proxy/Dockerfile`: multi-stage build — `golang:1.22` builder stage compiles static binary (`CGO_ENABLED=0`), `scratch` or `alpine` runtime stage copies binary
- [ ] 4.2 Build proxy image: `docker build -t openrouter-proxy spike/issue-45-opencode-cache/proxy/`
- [ ] 4.3 Verify image starts and exits cleanly when `OPENROUTER_API_KEY` is missing (should print error and exit 1)

## 5. Fixture Integration

- [ ] 5.1 Add custom provider entry to `spike/issue-45-opencode-cache/fixture/opencode.json` pointing at `http://openrouter-proxy:8080` with a placeholder token
- [ ] 5.2 Update `spike/issue-45-opencode-cache/README.md` with two-container startup instructions: create network, start proxy container with `OPENROUTER_API_KEY`, start fixture container on same network, run `opencode run` via `docker exec`

## 6. Validation

- [ ] 6.1 Start both containers on `spike-net`, run a single-turn `opencode run "Say: acknowledged"` through the proxy, confirm response received
- [ ] 6.2 Run a 3-turn session using `opencode run --session <id>` pattern, query `tokens_cache_read` from opencode DB, confirm cache behavior matches SF-3 baseline (512 tokens read on turn 2)
- [ ] 6.3 Close GitHub issue #55
