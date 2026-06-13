## Context

The devcontainer currently uses `devcontainer.json` with a `build.dockerfile` entry
targeting the `final` stage. Multi-stage Docker builds in this configuration only tag
the final image — intermediate stages are cached but not named. The `reinstall-node-tools`
change requires `symphony-maestro-node-builder` to be available by name on the host Docker
daemon so that `task node:build` and `task node:package:*` can run it via DooD.

Docker Compose is the supported devcontainer mechanism for building multiple named images
from a single Dockerfile.

## Goals / Non-Goals

**Goals:**
- `symphony-maestro-node-builder` available by name on the host Docker daemon after devcontainer build
- All existing devcontainer behaviour preserved (mounts, features, postStartCommand, remoteUser)
- ARG versions for node-builder stage passed through compose build args

**Non-Goals:**
- Running any services at container start (both services are build-only; no `depends_on` or runtime orchestration)
- Adding any new stages or modifying the Dockerfile
- Changing any task or post-start behaviour

## Decisions

### D1: docker-compose.yml with two build-only services

**Decision**: `.devcontainer/docker-compose.yml` defines two services — `devcontainer`
(`--target final`) and `node-builder` (`--target node-builder`, `image: symphony-maestro-node-builder`).
`devcontainer.json` switches from `build.dockerfile` to `dockerComposeFile` + `service: devcontainer`.
The `node-builder` service has no `command` — it is never started, only built.

**Rationale**: Docker Compose `build` with a named `image` field causes compose to tag the
built image with that name, making it available on the host Docker daemon. VS Code builds
all services in the compose file when opening the devcontainer, so `node-builder` is built
and tagged automatically alongside `final`.

**Alternative considered**: `postCreateCommand` running `docker build --target node-builder`.
Rejected: runs after the container is already up, introduces a network dependency at runtime,
and is not the intended devcontainer lifecycle hook for this purpose.

### D2: Build args for NODE_VERSION and PNPM_VERSION in compose

**Decision**: The `node-builder` service in `docker-compose.yml` passes `NODE_VERSION` and
`PNPM_VERSION` as build args with hardcoded defaults matching the Dockerfile ARGs. The
`devcontainer` service passes the same args for consistency (the base stage does not use
them but they are declared at the top of the Dockerfile).

**Rationale**: Compose build args must be explicitly listed to be passed through. Hardcoding
them in compose (mirroring the Dockerfile defaults) ensures they are always in sync — Renovate
updates the Dockerfile ARG defaults, and those are the single source of truth.

### D3: workspaceFolder and shutdownAction in devcontainer.json

**Decision**: When switching to `dockerComposeFile`, `workspaceMount` must be explicitly
specified in `devcontainer.json` because compose does not automatically bind-mount the
workspace. `shutdownAction` changes from `stopContainer` to `stopCompose`.

**Rationale**: Required by the devcontainer spec for compose-based configurations.
