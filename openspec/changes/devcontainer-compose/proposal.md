## Why

The `reinstall-node-tools` change requires a `node-builder` Docker image to be available
by name inside the devcontainer for use by DooD tasks. The current devcontainer uses
`devcontainer.json` with a single `build.dockerfile` entry targeting the `final` stage.
Multi-stage builds with this configuration do not produce named images for intermediate
stages — only the final image is tagged. This makes the `node-builder` stage inaccessible
to tasks that need to `docker run symphony-maestro-node-builder`.

Docker Compose is the supported mechanism for building multiple named images from the same
Dockerfile within a devcontainer. Switching `devcontainer.json` to use `dockerComposeFile`
with a `docker-compose.yml` that builds both the `final` (devcontainer) and `node-builder`
(tool runner) stages solves this cleanly without requiring any post-create hacks.

## What Changes

- Add `.devcontainer/docker-compose.yml` with two services:
  - `devcontainer`: `--target final`, the developer environment
  - `node-builder`: `--target node-builder`, tagged as `symphony-maestro-node-builder`
- Replace `build` block in `devcontainer.json` with `dockerComposeFile` and `service`
- All existing `mounts`, `features`, `postStartCommand`, `remoteUser`, and `customizations`
  are preserved unchanged

## Capabilities

### Modified Capabilities

- `devcontainer-environment`: devcontainer build mechanism changes from single-stage
  `dockerfile` to docker-compose multi-service build; `node-builder` image is now
  available by name on the host Docker daemon for DooD tasks

## Impact

- `.devcontainer/docker-compose.yml`: new file
- `.devcontainer/devcontainer.json`: `build` block replaced with `dockerComposeFile` + `service`
