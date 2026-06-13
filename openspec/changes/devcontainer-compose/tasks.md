## 1. docker-compose.yml

- [ ] 1.1 Create `.devcontainer/docker-compose.yml` with a `devcontainer` service (`build.target: final`, `image: symphony-maestro`) and a `node-builder` service (`build.target: node-builder`, `image: symphony-maestro-node-builder`); both services share the same `build.context` and `build.dockerfile`
- [ ] 1.2 Pass `NODE_VERSION` and `PNPM_VERSION` as build args in both services (values matching Dockerfile ARG defaults)

## 2. devcontainer.json

- [ ] 2.1 Replace the `build` block with `dockerComposeFile: "docker-compose.yml"` and `service: "devcontainer"`
- [ ] 2.2 Add explicit `workspaceMount` to bind-mount the workspace (required by devcontainer spec for compose configurations)
- [ ] 2.3 Change `shutdownAction` from `stopContainer` to `stopCompose`

## 3. Verification

- [ ] 3.1 Rebuild the devcontainer and confirm it opens successfully
- [ ] 3.2 Confirm `task devcontainer:doctor` passes inside the container
- [ ] 3.3 Confirm `docker image ls symphony-maestro-node-builder` shows the image from inside the devcontainer via DooD
- [ ] 3.4 Confirm all existing named volumes are still mounted (`vscode-extensions`, `vscode-user-data`)
