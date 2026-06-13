## 1. Proxy Changes

- [x] 1.1 Replace `io.Copy` streaming loop in `handleChatCompletion` with `io.ReadAll` to buffer full request body
- [x] 1.2 Replace `io.Copy` streaming loop for response with `io.ReadAll` to buffer full response body
- [x] 1.3 Add `logEntry` struct with `Timestamp`, `Request`, `Response`, `Error` fields (all JSON)
- [x] 1.4 Add `writeLog` function: marshal `logEntry` to indented JSON, write to `LOG_DIR/<timestamp>.json`
- [x] 1.5 Add `LOG_DIR` env var handling in `main`: default `/logs`, create dir on startup, log the path
- [x] 1.6 Call `writeLog` after each request: on success (request + response), on upstream error (request + error)
- [x] 1.7 Handle non-JSON response bodies: if `json.Valid` is false, store raw body as JSON string value

## 2. Rebuild and Verify

- [x] 2.1 Rebuild proxy image: `docker build -t openrouter-proxy spike/issue-45-opencode-cache/proxy/`
- [x] 2.2 Run 3-turn session with log volume mounted: `docker run ... -v /tmp/proxy-logs:/logs openrouter-proxy`
- [x] 2.3 Confirm log files exist and contain `request.messages` array and `response.usage.prompt_tokens_details`
- [x] 2.4 Confirm `messages[0].content` shape from live opencode traffic (plain string vs. content-parts array)

## 3. Documentation

- [x] 3.1 Update `proxy/AGENTS.md`: add `LOG_DIR` env var row and volume mount example
- [x] 3.2 Update `README.md` proxy run command to include `-v /tmp/proxy-logs:/logs` example
- [x] 3.3 Write `findings/sf-4b-proxy-inspection.md`: document `messages[0].content` shape, `prompt_tokens_details` field presence, and any other wire-format findings

## 4. Close Out

- [x] 4.1 Commit and push on `feature/issue-55-openrouter-proxy`
- [x] 4.2 Close GitHub issue #56 — `Closes #56` in commit message; will close on PR #59 merge
