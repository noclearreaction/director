## Context

Symphony-maestro currently has no devcontainer definition. Developers and agent sessions depend on host-installed tooling — `deno`, `go`, `openspec`, `opencode`, `gh` — at whatever versions happen to be installed. The sibling repos `poc-bdd-tdd` and `wip-devops-copilot-lab2` both have mature multi-stage devcontainer patterns that inform this design.

The repo uses Deno/TypeScript for pipeline tooling (`bin/`), Go for core code, and is forced to accept Node/pnpm as a runtime dependency of `openspec` and `opencode` (external CLIs with no standalone binary alternative today).

## Goals / Non-Goals

**Goals:**
- Reproducible environment: all required tools installed at pinned versions inside the container
- Multi-stage Dockerfile separating download, runtime assembly, and user environment stages
- Linux FHS compliance for versioned runtimes: `/opt/<tool><version>/` with `update-alternatives` managing `/usr/local/bin/` entries
- A `ci` stage usable by future CI pipelines, distinct from the developer-facing `final` stage
- Node/pnpm available but not on the dev `PATH` — only `openspec` and `opencode` wrappers exposed
- DooD (Docker-outside-of-Docker) available for future sidecar needs, invoked from entrypoint/postStart rather than Compose

**Non-Goals:**
- Docker Compose: deferred; DooD handles sidecar needs without it
- `docker buildx bake`: creates a host tooling dependency — the same problem devcontainers solve; rejected
- Shared `/opt` volume across devcontainers: noted as a future possibility the `update-alternatives` layout enables, deferred
- Moving the base image to a separate repository: the Dockerfile stages are structured to make this easy later, but the split is deferred
- Converting `bin/` scripts to use a project `deno.json`: tracked separately, handled within this change's task list

## Decisions

### D1: Plain `docker build` multi-stage — not Compose, not bake

**Decision**: Use a plain multi-stage `Dockerfile` with `devcontainer.json` pointing at it via `build.dockerfile`. No `docker-compose.yaml`, no `docker buildx bake`.

**Rationale**: `bake` requires buildx installed on every host that opens the repo — a host dependency that is exactly the antipattern devcontainers solve. Compose adds lifecycle complexity and has historically caused friction; DooD sidecars don't require it. Multi-stage Dockerfiles provide the same stage separation that bake offers without any host tooling requirement.

**Future path**: when the base image moves to a dedicated repo, the `FROM` line in one stage changes to `FROM ghcr.io/org/dev-base:latest` — no other files change.

---

### D2: Go installed under `/opt/go<version>/` — not `/usr/local/go`

**Decision**: Go is copied from the `go-runtime` build stage into `/opt/go<version>/`. `GOROOT` points at the versioned path. `update-alternatives` manages `/usr/local/bin/go` with `gofmt` as a slave.

**Rationale**: `/usr/local/go` is the Go installer convention but conflates version identity with location. Versioned paths under `/opt/` follow Linux FHS for locally-managed software and make future side-by-side installs or the shared `/opt` volume scenario tractable without filesystem surgery.

**Alternatives considered**: `/usr/local/go` (conventional, rejected — version unidentifiable from path alone); `/usr/lib/go-<version>` (Debian package convention, rejected — we're not using apt for Go).

---

### D3: `update-alternatives` for all versioned binaries

**Decision**: Go (`/usr/local/bin/go`, slave: `gofmt`) and Deno (`/usr/local/bin/deno`) are registered via `update-alternatives`. Task is a single-use binary with no meaningful version coexistence need — installed directly to `/usr/local/bin/task`.

**Rationale**: `update-alternatives` is the Debian standard for managing multiple versions of the same tool. Slaving `gofmt` to `go` ensures they move atomically. This also lays the groundwork for the deferred shared `/opt` volume idea.

---

### D4: Node/pnpm topology — not on `PATH`

**Decision**: Node is installed to `/opt/node/`. Only `node` binary is on `PATH` (required by openspec/opencode wrappers). `pnpm`, `npm`, `npx`, and `corepack` are intentionally absent from `PATH`. No wrapper scripts that intercept and block — topology enforces the constraint.

**Rationale**: Wrapper interception scripts (e.g. `bin/npm` that prints an error) are fragile and non-standard. Removing these tools from `PATH` is the correct standard approach. Developers who genuinely need pnpm can use the full path; this being inconvenient is intentional.

---

### D5: `ci` stage sits between `base` and `final`

**Decision**: Stage order: `download-base` → runtime stages → `base` → `ci` → `final`.

`ci` inherits `base` and adds only what a non-interactive automated runner needs (clean ENV, no user state).

`final` inherits `ci` and adds: SSH agent profile scripts, `~/.bashrc`/`~/.zshrc` setup, vscode user directory prep, git hook wiring, `env-lgc` tag, XDG env vars.

**Rationale**: CI should be a strict subset of the dev environment. Making `ci` an explicit stage that `final` extends ensures they can never diverge silently.

---

### D6: Task installed from standalone go-task GitHub release

**Decision**: Task binary downloaded from `https://github.com/go-task/task/releases` in a dedicated `task-binary` stage. Not installed via npm (`@go-task/cli`).

**Rationale**: Consistent with the principle of keeping Node/pnpm constrained. The standalone binary is the upstream-preferred distribution method for non-Node environments.

---

### D7: DooD for sidecars

**Decision**: Include `ghcr.io/devcontainers/features/docker-outside-of-docker:1` in devcontainer features. Sidecars (future: MCP proxy, test services) are started from `postStartCommand` or entrypoint scripts using `docker run`, not from a `docker-compose.yaml`.

**Rationale**: DooD is simpler to reason about than Compose for devcontainer sidecar needs. The Docker socket is available; containers started this way live and die with the dev session naturally.

## Risks / Trade-offs

- **pnpm/Node accepted as necessary evil** → Mitigation: isolated in `/opt/node/`, not on PATH, single-purpose wrappers for openspec/opencode only. If openspec ships a standalone binary in future, Node can be removed entirely.
- **Hardcoded Deno shebang paths in `bin/`** → `bin/commit-lint.ts` and `bin/provision-labels.ts` use `/home/tunnel49/.deno/bin/deno` — this path won't exist in the container. Mitigation: update shebangs to use `/usr/bin/env deno` within this change.
- **No rebuild trigger on `ARG` version changes** → VS Code doesn't auto-rebuild when Dockerfile ARGs change. Mitigation: documented in `.devcontainer/` README note; developers rebuild manually with "Rebuild Container".
- **update-alternatives adds `PATH` indirection for GOROOT** → GOROOT must point at `/opt/go<version>/` not at the alternatives symlink, since Go needs the full tree. Mitigation: set `GOROOT` explicitly in ENV to the versioned path.

## Open Questions

- Does `postStartCommand` need to wire git hooks (`.devcontainer/setup-mcp.sh` equivalent), and if so, what does that script contain for this repo specifically?

---

### D8: All tool versions pinned; update-flagging deferred

**Decision**: Every tool installed in the Dockerfile SHALL have an explicit `ARG <TOOL>_VERSION` with a pinned value. No tool is installed as `latest` or `@latest` at build time. This applies to: `GO_VERSION`, `DENO_VERSION`, `TASK_VERSION`, `NODE_VERSION`, `PNPM_VERSION`, `OPENSPEC_VERSION`, `OPENCODE_VERSION`, and any future additions.

A mechanism to flag available updates (e.g. Renovate, a Deno script that checks GitHub releases, or a Task target) is explicitly a future concern and out of scope for this change. The version ARG layout — all versions in one place at the top of the Dockerfile — is designed to make such a system easy to add.

**Rationale**: Unpinned versions make builds non-reproducible and introduce silent breakage. The update-flagging system is a separate concern that requires its own design (what triggers it, where it reports, how it integrates with the change workflow). Blocking this change on that design would be premature.

**Non-goal marker**: A future change should own `devcontainer-version-updates` as a capability — scanning ARG values against upstream releases and surfacing a diff or PR.
