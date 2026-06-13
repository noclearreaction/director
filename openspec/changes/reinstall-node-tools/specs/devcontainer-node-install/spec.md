## ADDED Requirements

### Requirement: npm packages installed into named Docker volumes at container start
Node packages SHALL NOT be installed at Docker image build time. Two named Docker volumes SHALL be declared in `devcontainer.json`:
- `${localWorkspaceFolderBasename}-node-modules` mounted at `/opt/node/node_modules` — the deployed package tree, populated by the builder
- `${localWorkspaceFolderBasename}-pnpm-store` mounted at `/root/.local/share/pnpm/store` inside builder containers — the persistent pnpm content store

A `task node:build` task SHALL run the `symphony-maestro-node-builder` image via DooD: wipe the volume's `node_modules`, run `pnpm install --frozen-lockfile`, run `pnpm deploy /dest` (where `/dest/node_modules` is the volume), then run `pnpm store prune`. `task node:build` SHALL be called in `post-start` on every container start.

#### Scenario: Volume populated on container start
- **WHEN** the devcontainer starts
- **THEN** `task node:build` runs and populates the `node-modules` volume
- **AND** openspec, opencode, and renovate are available on PATH

#### Scenario: Package added without image rebuild
- **WHEN** a developer runs `task node:package:add -- foo 1.0.0` followed by `task node:build`
- **THEN** `foo` is available in the devcontainer without rebuilding the image

#### Scenario: Volume shared between devcontainer and DooD builder
- **WHEN** the builder container writes to the `node-modules` volume via DooD
- **THEN** the devcontainer reads the updated packages from the same volume

---

### Requirement: npm packages installed from committed lockfile
The builder container SHALL use a committed `pnpm-lock.yaml` and run `pnpm install --frozen-lockfile`. The install SHALL fail if the lockfile does not match `package.json`.

#### Scenario: Builder fails when lockfile is stale
- **WHEN** `package.json` is updated without regenerating `pnpm-lock.yaml`
- **THEN** `task node:build` fails with a frozen-lockfile mismatch error

#### Scenario: Builder succeeds with matching lockfile
- **WHEN** `package.json` and `pnpm-lock.yaml` are in sync
- **THEN** `pnpm install --frozen-lockfile` succeeds and installs exact versions from the lockfile

---

### Requirement: pnpm deploy produces portable node_modules in the volume
The builder container SHALL run `pnpm deploy` to produce a real-files-only node_modules directory written to the `node-modules` volume. The deploy output SHALL contain no symlinks to the pnpm content store.

#### Scenario: Deployed packages load without store symlinks
- **WHEN** the final image is run without the pnpm content store present
- **THEN** `require()` on deployed packages succeeds

#### Scenario: Native module artifacts present in deployed output
- **WHEN** a package with an install script (e.g. opencode-ai) is included
- **THEN** the build artifacts produced by the install script are present in the deploy output

---

### Requirement: Build script approvals are committed and reviewed
The committed `pnpm-workspace.yaml` file SHALL be committed to the repo and SHALL contain the pnpm supply chain hardening settings. It is the sole file modified by `task node:trust:add` and `task node:trust:rm` to manage build script approvals. Its internal format is owned by pnpm.

#### Scenario: Unapproved package install script is blocked
- **WHEN** a transitive dependency with an install script is not listed in `allowBuilds`
- **THEN** pnpm does not run the script and the install completes without error

#### Scenario: Approved package install script runs
- **WHEN** a package is listed in `allowBuilds: true`
- **THEN** pnpm runs its install script as part of `pnpm install`

---

### Requirement: pnpm supply chain hardening settings applied
The committed `pnpm-workspace.yaml` SHALL set `minimumReleaseAge: 10080` (7 days), `blockExoticSubdeps: true`, and `trustPolicy: no-downgrade`.

#### Scenario: Package published less than 7 days ago is not resolved
- **WHEN** a package version was published fewer than 10080 minutes ago
- **THEN** pnpm does not resolve it during `pnpm install` (outside of frozen lockfile mode)

#### Scenario: Exotic transitive dependency is blocked
- **WHEN** a transitive dependency references a git URL or direct tarball
- **THEN** `pnpm install` fails with an exotic subdep error

---

### Requirement: npm packages managed via sandboxed task interface
Npm packages SHALL be managed exclusively via `task node:package:add`, `task node:package:rm`, `task node:package:list`, `task node:package:update`, `task node:package:audit`, and `task node:package:prune`. Each mutating task SHALL execute pnpm in a throwaway Docker container (sandboxed: no host filesystem access beyond `.devcontainer/node/`, exits immediately after completion). The tasks SHALL update `package.json` and `pnpm-lock.yaml` atomically. `task node:package:list` SHALL print the current direct dependencies without running a container. pnpm SHALL NOT be installed in the devcontainer image or available on PATH.

#### Scenario: Adding a new npm package
- **WHEN** a developer runs `task node:package:add -- renovate 43.220.0`
- **THEN** the package is added to `package.json` and `pnpm-lock.yaml` is regenerated
- **AND** no entry is added to `allowBuilds` in `pnpm-workspace.yaml`

#### Scenario: Updating an existing npm package
- **WHEN** a developer runs `task node:package:update -- opencode-ai 1.17.0`
- **THEN** the version is updated in `package.json` and `pnpm-lock.yaml` is regenerated

#### Scenario: Removing an npm package
- **WHEN** a developer runs `task node:package:rm -- opencode-ai`
- **THEN** the package is removed from `package.json` and `pnpm-lock.yaml` is regenerated
- **AND** if the package had an `allowBuilds` entry it is removed from `pnpm-workspace.yaml`

#### Scenario: Listing installed packages
- **WHEN** a developer runs `task node:package:list`
- **THEN** the current direct dependencies from `package.json` are printed
- **AND** no Docker container is started

#### Scenario: Auditing for known vulnerabilities
- **WHEN** a developer runs `task node:package:audit`
- **THEN** `pnpm audit` is run in a sandboxed container against the current lockfile and results are printed
- **AND** no mutations are made to any file

#### Scenario: Pruning the pnpm store
- **WHEN** a developer runs `task node:package:prune`
- **THEN** `pnpm store prune` is run in a builder container with the pnpm store volume mounted
- **AND** store entries not referenced by the current lockfile are removed
- **AND** no changes are made to `package.json` or `pnpm-lock.yaml`

#### Scenario: pnpm not available in devcontainer
- **WHEN** a developer runs `pnpm` inside the devcontainer
- **THEN** the shell returns "command not found"

#### Scenario: Sandbox has no access beyond .devcontainer/node/
- **WHEN** any mutating `task node:*` task runs
- **THEN** the Docker container has no mount access to any path outside `.devcontainer/node/`

---

### Requirement: Build script trust managed independently of package installation
Approving or revoking a package's build script permission SHALL be done via `task node:trust:add`, `task node:trust:rm`, and `task node:trust:list`, independently of whether the package is a direct or transitive dependency. This is necessary because transitive deps with build scripts (e.g. re2 as a transitive dep of renovate) require explicit approval but are never added via `task node:package:add`.

#### Scenario: Trusting a transitive dependency's build script
- **WHEN** a developer runs `task node:trust:add -- re2`
- **THEN** `re2: true` is set in `allowBuilds` in `pnpm-workspace.yaml`
- **AND** no changes are made to `package.json` or `pnpm-lock.yaml`

#### Scenario: Revoking build script trust
- **WHEN** a developer runs `task node:trust:rm -- re2`
- **THEN** the `re2` entry is removed from `allowBuilds` in `pnpm-workspace.yaml`

#### Scenario: Listing trusted packages
- **WHEN** a developer runs `task node:trust:list`
- **THEN** the current `allowBuilds` entries from `pnpm-workspace.yaml` are printed

#### Scenario: Verifying trust coverage before docker build
- **WHEN** a developer runs `task node:trust:verify`
- **THEN** pnpm is run in a sandboxed container to identify any transitive deps with build scripts not yet listed in `allowBuilds`
- **AND** the task exits non-zero if any unapproved build scripts are found
- **AND** no mutations are made to any file

#### Scenario: Removing a direct package cleans up its trust entry
- **WHEN** a developer runs `task node:package:rm -- opencode-ai`
- **THEN** the `opencode-ai` entry is removed from `allowBuilds` if present
