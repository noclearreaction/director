## MODIFIED Requirements

### Requirement: Devcontainer built via docker-compose with named service images
The devcontainer SHALL be defined using `dockerComposeFile` in `devcontainer.json` pointing
to `.devcontainer/docker-compose.yml`. The compose file SHALL define at minimum two services
built from the same Dockerfile:
- `devcontainer`: `--target final`, used as the developer environment
- `node-builder`: `--target node-builder`, tagged as `symphony-maestro-node-builder`

Both services SHALL be built when the devcontainer is built. The `node-builder` image SHALL
be available by name on the host Docker daemon after a successful build.

#### Scenario: node-builder image available after devcontainer build
- **WHEN** VS Code builds the devcontainer
- **THEN** the `symphony-maestro-node-builder` image is present on the host Docker daemon
- **AND** `docker run --rm symphony-maestro-node-builder pnpm --version` succeeds from inside the devcontainer via DooD

#### Scenario: Devcontainer environment unchanged
- **WHEN** the devcontainer starts after the compose migration
- **THEN** all existing tools (Go, Deno, Task, gh, Docker) remain available
- **AND** `task devcontainer:doctor` passes
- **AND** all existing named volumes (`vscode-extensions`, `vscode-user-data`) remain mounted

#### Scenario: ARG versions passed through compose to Dockerfile
- **WHEN** the compose file specifies `NODE_VERSION` and `PNPM_VERSION` build args
- **THEN** the `node-builder` stage uses exactly those versions
