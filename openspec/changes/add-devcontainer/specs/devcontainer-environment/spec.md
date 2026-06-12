## ADDED Requirements

### Requirement: Multi-stage Dockerfile assembles all runtimes
The devcontainer SHALL be built from a single multi-stage `Dockerfile` with no external host tooling prerequisites beyond `docker build`. Stages SHALL be: `download-base`, `go-runtime`, `deno-runtime`, `task-binary`, `node-runtime`, `base`, `ci`, `final`.

#### Scenario: Container builds from clean Docker cache
- **WHEN** a developer runs `docker build --target final .devcontainer/`
- **THEN** the build succeeds, producing an image with Go, Deno, Task, `openspec`, `opencode`, and `gh` available

#### Scenario: CI target builds independently
- **WHEN** a CI runner builds with `--target ci`
- **THEN** the image contains Go, Deno, Task, `openspec`, `opencode`, and `gh`, but excludes SSH profile scripts and vscode user directory setup

---

### Requirement: Go installed under versioned FHS path with update-alternatives
Go SHALL be installed at `/opt/go<VERSION>/` where `<VERSION>` matches the `GO_VERSION` ARG. `update-alternatives` SHALL register `/usr/local/bin/go` pointing at `/opt/go<VERSION>/bin/go`, with `gofmt` registered as a slave alternative. `GOROOT` ENV SHALL be set to the versioned path (e.g. `/opt/go1.26.4`), not to a symlink.

#### Scenario: Go binary reachable on PATH
- **WHEN** a shell session starts in the container
- **THEN** `go version` outputs the expected version and `gofmt --help` succeeds

#### Scenario: gofmt slave follows go alternative
- **WHEN** `update-alternatives --set go /opt/go<VERSION>/bin/go` is called
- **THEN** `update-alternatives --display gofmt` shows the corresponding `/opt/go<VERSION>/bin/gofmt` as the active slave

---

### Requirement: Deno installed under versioned FHS path with update-alternatives
Deno SHALL be installed at `/opt/deno-<VERSION>/bin/deno`. `update-alternatives` SHALL register `/usr/local/bin/deno` pointing at that path.

#### Scenario: Deno binary reachable on PATH
- **WHEN** a shell session starts in the container
- **THEN** `deno --version` outputs the expected version

---

### Requirement: Task installed from standalone binary
Task SHALL be installed directly to `/usr/local/bin/task` from the official go-task GitHub release archive. It SHALL NOT be installed via npm or pnpm.

#### Scenario: Task runs without Node
- **WHEN** `task --version` is run in the container
- **THEN** the command succeeds and the version matches the pinned `TASK_VERSION` ARG

---

### Requirement: Node and pnpm isolated from dev PATH
Node SHALL be installed to `/opt/node/`. Only the `node` binary SHALL be on `PATH`. `pnpm`, `npm`, `npx`, and `corepack` SHALL NOT appear on `PATH`. `openspec` and `opencode` CLI wrappers SHALL be the only pnpm-installed tools exposed in `/usr/local/bin/`.

#### Scenario: pnpm not reachable by default
- **WHEN** a developer runs `pnpm` in a container shell
- **THEN** the shell returns "command not found"

#### Scenario: openspec reachable without pnpm on PATH
- **WHEN** a developer runs `openspec --version`
- **THEN** the command succeeds

#### Scenario: opencode reachable without pnpm on PATH
- **WHEN** a developer runs `opencode --version`
- **THEN** the command succeeds

---

### Requirement: bin/ scripts use portable Deno shebang
All scripts in `bin/` that invoke Deno SHALL use `#!/usr/bin/env deno` (or `#!/usr/bin/env -S deno run ...`) rather than hardcoded host paths (e.g. `/home/tunnel49/.deno/bin/deno`).

#### Scenario: commit-lint runs inside container
- **WHEN** `bin/commit-lint.ts` is executed inside the devcontainer
- **THEN** it invokes Deno successfully without a "No such file or directory" error

#### Scenario: provision-labels runs inside container
- **WHEN** `bin/provision-labels.ts` is executed inside the devcontainer
- **THEN** it invokes Deno successfully without a "No such file or directory" error

---

### Requirement: devcontainer.json wires VS Code to the Dockerfile build
A `devcontainer.json` SHALL reference the Dockerfile via `build.dockerfile` targeting the `final` stage. It SHALL include the `docker-outside-of-docker` devcontainer feature. It SHALL mount named volumes for VS Code server extensions and user data, consistent with the existing sibling-repo pattern.

#### Scenario: VS Code opens in container
- **WHEN** a developer opens the repository in VS Code with the Dev Containers extension
- **THEN** VS Code detects the devcontainer configuration and offers to reopen in container

#### Scenario: Docker available inside container
- **WHEN** a developer runs `docker ps` inside the container
- **THEN** the command succeeds, communicating with the host Docker daemon via DooD

---

### Requirement: All tool versions explicitly pinned in Dockerfile
Every tool installed in the Dockerfile SHALL have a corresponding `ARG <TOOL>_VERSION` declared at the top of the file with a concrete pinned value. No tool SHALL be installed using `latest`, `@latest`, or any floating version reference at build time.

#### Scenario: Dockerfile ARGs enumerate all tool versions
- **WHEN** a developer reads the top of the Dockerfile
- **THEN** they find explicit ARG declarations for GO_VERSION, DENO_VERSION, TASK_VERSION, NODE_VERSION, PNPM_VERSION, OPENSPEC_VERSION, and OPENCODE_VERSION, each with a concrete pinned value

#### Scenario: No floating version references in RUN instructions
- **WHEN** the Dockerfile is scanned for `@latest` or unversioned install commands
- **THEN** none are found in any `RUN` instruction that installs a tool
