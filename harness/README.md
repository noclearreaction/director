# opencode cache harness

A minimal, reproducible Docker environment for exploring opencode cache behavior as part of spike [#43](https://github.com/noclearreaction/symphony-director/issues/43).

This harness is the baseline for SF-1 ([#45](https://github.com/noclearreaction/symphony-director/issues/45)). It provides a clean, isolated container with opencode installed and a minimal project fixture ready to use.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) installed and running

## Pinned versions

| Software | Version |
|---|---|
| Node.js (base image) | `node:20-slim` |
| opencode | `1.16.2` |

To update the opencode version, edit the `RUN npm install -g opencode-ai@...` line in `Dockerfile` and rebuild.

## Fixture layout

The minimal project fixture is baked into the image at `/app/fixture/`:

```
/app/fixture/
├── opencode.json   # opencode config: loads AGENTS.md, sets default agent
└── AGENTS.md       # ~160-token system prompt for the experiment agent
```

The fixture defines a single agent (`experiment`) with a short, low-variability system prompt. It has no application code and no tools configured.

## Build the image

```bash
docker build -t opencode-cache-harness harness/
```

Run from the repository root. The build installs opencode globally inside the image.

## Start an experiment session

```bash
docker run --rm -it opencode-cache-harness
```

This drops you into a bash shell inside the container with:
- `opencode` available on the PATH
- Working directory set to `/app/fixture/` (the minimal project)

From there you can invoke opencode directly and explore its behavior:

```bash
# Check opencode is available
opencode --version

# Explore available commands
opencode --help

# Check what database tooling is available
opencode db --help

# View session stats
opencode stats
```

## Extending the fixture

To iterate on the system prompt without rebuilding:

```bash
docker run --rm -it \
  -v "$(pwd)/harness/fixture:/app/fixture" \
  opencode-cache-harness
```

This volume-mounts your local fixture over the baked-in one, so edits are reflected immediately without a rebuild. Use this for prompt iteration; the baked-in fixture is the reproducible baseline.

## Notes

- The opencode version is pinned in the Dockerfile. Do not change it mid-spike without documenting the change as a variable.
- The `experiment` agent system prompt is approximately 160 tokens. This is an approximation — actual token count depends on the model's tokenizer.
- How to trigger turns, where token usage is recorded, and how to read cache metrics are open questions this spike is designed to answer.
