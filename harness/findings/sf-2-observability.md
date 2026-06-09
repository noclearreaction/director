# SF-2 Observability Findings

Discovery session for issue #46 (`explore-opencode-debug-log-structure`).  
opencode version: **1.16.2** | Model used: **opencode/deepseek-v4-flash-free** (no credentials required)  
Harness image: **opencode-cache-harness:latest** (from SF-1 / #45)

---

## CLI Subcommands

Full output of `opencode --help`. Key subcommands for measurement:

| Subcommand | Description |
|---|---|
| `opencode run [message]` | Run opencode with a message (non-interactive). Response goes to TUI/stderr only — **stdout is empty**. |
| `opencode db [query]` | Execute a SQL query against the SQLite database or open interactive shell. `--format json\|tsv`. |
| `opencode db path` | Print the database file path. |
| `opencode stats` | Show human-readable token usage and cost statistics (all-time or `--days N`). |
| `opencode export [sessionID]` | Export a session as structured JSON including messages, parts, and token counts. |
| `opencode session` | Manage sessions (list, etc.). |
| `opencode models` | List available models. |

Other subcommands present: `completion`, `acp`, `mcp`, `attach`, `debug`, `providers`, `agent`, `upgrade`, `uninstall`, `serve`, `web`, `github`, `pr`, `plugin`.

---

## Debug Log

Captured with `opencode run --print-logs --log-level DEBUG "..." 2>debug.log`.

### Structure

The debug log is **line-oriented**, one JSON-like entry per line with format:
```
LEVEL  TIMESTAMP +ELAPSEDms service=<name> [key=value ...] <message>
```

### Phases observed (in order)

1. **Startup** — `service=default`: version, args, process_role, run_id
2. **Instance creation** — `service=default`: directory
3. **Project init** — `service=project`: directory
4. **Config loading** — `service=config`: loads from `~/.config/opencode/config.json`, `opencode.json`, `opencode.jsonc`, then project-local `opencode.json`
5. **Plugin loading** — `service=plugin`: loads ~9 internal plugins by short name (obfuscated: `sW`, `JM`, `V7`, etc.)
6. **LSP / formatter init** — disabled in this fixture
7. **Session created** — `service=session`: full session state as structured log fields (see below)
8. **Server event** — `service=server`: `event=connected`
9. **Provider init** — `service=provider`: `providerID=opencode`, `status=started/completed`, duration
10. **Session prompt loop** — `service=session.prompt`: `step=0`, `step=1`, then `exiting loop`
11. **Tool resolution** — `service=session.tools` + `service=tool.registry`: registers bash, read, glob, grep, edit, write, task, webfetch, todowrite, websearch, skill
12. **LLM call** — `service=llm`: `providerID=opencode`, `modelID=deepseek-v4-flash-free`, `session.id`, `agent=build`, `mode=primary`, `stream`
13. **Instance dispose** — `service=default`

### Session creation log entry (key fields)

```
service=session
  id=ses_...          # session ID
  slug=crisp-cabin    # human-readable slug
  version=1.16.2
  projectID=global
  directory=/app/fixture
  path=app/fixture
  title="New session - <ISO timestamp>"
  permission=[...]    # JSON array
  cost=0
  tokens={"input":0,"output":0,"reasoning":0,"cache":{"read":0,"write":0}}
  time={"created":<unix_ms>,"updated":<unix_ms>}
```

### Token/cache fields in the debug log

Token counts in the debug log are **session-level running totals** logged at session creation only (all zeros at that point). They are **not** logged per-turn or after the LLM call completes. **The debug log does not contain post-turn token counts.**

---

## Database Schema

Database path: `/root/.local/share/opencode/opencode.db`  
Accessible via: `opencode db "<SQL>"` or direct `sqlite3` query.

### All tables

`migration`, `project`, `message`, `part`, `session`, `todo`, `session_share`, `control_account`, `account`, `account_state`, `event_sequence`, `event`, `workspace`, `session_message`, `data_migration`, `permission`, `project_directory`, `sqlite_sequence`, `session_input`, `session_context_epoch`

### `session` table — key columns

| Column | Type | Notes |
|---|---|---|
| `id` | TEXT (PK) | e.g. `ses_1519e289cffe...` |
| `project_id` | TEXT | |
| `slug` | TEXT | human-readable, e.g. `crisp-squid` |
| `directory` | TEXT | working directory |
| `title` | TEXT | auto-generated |
| `agent` | TEXT | e.g. `build` |
| `model` | TEXT | JSON: `{"id":"deepseek-v4-flash-free","providerID":"opencode","variant":"default"}` |
| `cost` | REAL | session-level cost total (0.0 for free models) |
| `tokens_input` | INTEGER | cumulative input tokens for the session |
| `tokens_output` | INTEGER | cumulative output tokens |
| `tokens_reasoning` | INTEGER | reasoning tokens (e.g. 20 observed) |
| `tokens_cache_read` | INTEGER | cumulative cache read tokens |
| `tokens_cache_write` | INTEGER | cumulative cache write tokens |
| `time_created` | INTEGER | Unix ms |
| `time_updated` | INTEGER | Unix ms |

### `message` table

| Column | Type | Notes |
|---|---|---|
| `id` | TEXT (PK) | |
| `session_id` | TEXT | FK → session |
| `time_created` | INTEGER | |
| `time_updated` | INTEGER | |
| `data` | TEXT | JSON blob containing role, agent, model, summary |

### `part` table

| Column | Type | Notes |
|---|---|---|
| `id` | TEXT (PK) | |
| `message_id` | TEXT | FK → message |
| `session_id` | TEXT | |
| `time_created` | INTEGER | |
| `time_updated` | INTEGER | |
| `data` | TEXT | JSON blob: `{"type":"text","text":"..."}` or tool call parts |

### Example session row (real values)

```json
{
  "id": "ses_1519cca48ffeEgwW1NL1CkbX26",
  "model": "{\"id\":\"deepseek-v4-flash-free\",\"providerID\":\"opencode\",\"variant\":\"default\"}",
  "cost": 0,
  "tokens_input": 8221,
  "tokens_output": 2,
  "tokens_reasoning": 20,
  "tokens_cache_read": 0,
  "tokens_cache_write": 0
}
```

Note: 8221 input tokens for a minimal "What is 2+2?" prompt — this reflects the large built-in system prompt from the `build` agent, not the user message alone.

---

## `opencode export` JSON structure

`opencode export <sessionID>` returns a JSON object with:

```json
{
  "info": {
    "id": "ses_...",
    "slug": "...",
    "agent": "build",
    "model": { "id": "deepseek-v4-flash-free", "providerID": "opencode", "variant": "default" },
    "cost": 0,
    "tokens": {
      "input": 8221, "output": 2, "reasoning": 20,
      "cache": { "read": 0, "write": 0 }
    },
    "time": { "created": 1781042211654, "updated": 1781042211900 }
  },
  "messages": [
    {
      "info": { "role": "user", "id": "msg_...", "sessionID": "ses_..." },
      "parts": [ { "type": "text", "text": "...", "id": "prt_..." } ]
    }
  ]
}
```

---

## Recommended Measurement Approach

For SF-3–SF-8, use **`opencode db` SQL queries** against the `session` table:

```bash
# Get token counts and cost for the most recent session
opencode db "SELECT tokens_input, tokens_output, tokens_reasoning, tokens_cache_read, tokens_cache_write, cost FROM session ORDER BY time_created DESC LIMIT 1" --format json
```

For per-session tracking across multiple runs, query by session ID or use `time_created` ordering.

For a human-readable summary, `opencode stats` works but is not machine-parseable.

For full message/response content, use `opencode export <sessionID>`.

**Note on token granularity**: The database stores session-level **cumulative** totals, not per-turn deltas. For multi-turn sessions, calculate deltas between successive queries or use per-turn `opencode export` data.

---

## Gaps vs #43 Assumptions

| #43 Assumption | Reality | Verdict |
|---|---|---|
| `tokens_input` field exists in db | ✅ Exact column name in `session` table | **Correct** |
| `tokens_cache_read` field exists | ✅ Exact column name in `session` table | **Correct** |
| `cost` field exists | ✅ `cost` REAL column in `session` table | **Correct** |
| `opencode db` command exists | ✅ Exists, accepts SQL, `--format json\|tsv` | **Correct** |
| `run-turn.sh` to trigger a turn | ⚠️ `opencode run "<message>"` works, but stdout is empty — response is TUI/stderr only. A wrapper script can redirect stderr but the text response is not in stdout. | **Partial — approach needs refinement** |
| `extract-metrics.sh` with `opencode db` | ✅ `opencode db "<SQL>" --format json` works exactly as intended | **Correct** |
| Token counts in debug log | ❌ Post-turn token counts are **not** in the debug log. They are only in the database. | **Incorrect — use DB, not log** |
| Cache read tokens observable | ✅ In DB as `tokens_cache_read`, but **0 for free models** — Anthropic cache behavior may require Anthropic models | **DB column exists; free model shows 0** |

### Key finding on cache tokens

`tokens_cache_read` and `tokens_cache_write` are **0** for `opencode/deepseek-v4-flash-free`. Cache behavior (prompt caching) is a provider-specific feature. To observe non-zero cache values, experiments likely need an Anthropic model with prompt caching enabled. This should be confirmed in SF-3.
