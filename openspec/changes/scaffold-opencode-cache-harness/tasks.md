## 1. Repository Structure

- [ ] 1.1 Create `harness/` directory at the repository root
- [ ] 1.2 Create `harness/fixture/` subdirectory for the minimal project files

## 2. Docker Image

- [ ] 2.1 Write `harness/Dockerfile` using `node:20-slim` as the base image
- [ ] 2.2 Pin the opencode version in the Dockerfile (`npm install -g opencode-ai@<version>`)
- [ ] 2.3 Verify `docker build -t opencode-cache-harness harness/` succeeds without errors

## 3. Minimal Project Fixture

- [ ] 3.1 Create `harness/fixture/agent.json` (single agent configuration)
- [ ] 3.2 Write the agent system prompt to `harness/fixture/system-prompt.txt` (~200 tokens, no application code)
- [ ] 3.3 Confirm fixture files are copied into the image at a documented path (e.g., `/app/fixture/`)

## 4. End-to-End Validation

- [ ] 4.1 Start a container interactively and confirm opencode is on the PATH
- [ ] 4.2 Confirm the fixture project is present at the documented working directory inside the container
- [ ] 4.3 Confirm opencode can be invoked manually (e.g., `opencode --help` or equivalent)

## 5. Documentation

- [ ] 5.1 Write `harness/README.md` covering prerequisites (Docker), `docker build` command, and how to start an interactive experiment session
- [ ] 5.2 Document the pinned opencode version and the fixture layout in the README
