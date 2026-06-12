# =============================================================================
# VERSION CONFIGURATION — UPDATE THESE TO UPGRADE TOOLS
# =============================================================================
# Go toolchain — https://golang.org/dl/
ARG GO_VERSION=1.26.4
# Deno — https://github.com/denoland/deno/releases
ARG DENO_VERSION=2.8.2
# Task (go-task) — https://github.com/go-task/task/releases
ARG TASK_VERSION=3.51.1
# Node.js — https://nodejs.org/en/download/releases/
ARG NODE_VERSION=24.16.0
# pnpm — https://github.com/pnpm/pnpm/releases
ARG PNPM_VERSION=11.5.2
# openspec (@fission-ai/openspec) — https://www.npmjs.com/package/@fission-ai/openspec
ARG OPENSPEC_VERSION=1.4.1
# opencode (opencode-ai) — https://www.npmjs.com/package/opencode-ai
ARG OPENCODE_VERSION=1.16.2

# =============================================================================
# Stage: download-base
# Purpose: Common utilities for binary downloads
# Output: /srv/files/ working directory with curl, ca-certificates, tar, xz-utils
# =============================================================================
FROM debian:bookworm-slim AS download-base
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
        curl \
        ca-certificates \
        tar \
        xz-utils \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /srv/files

# =============================================================================
# Stage: go-runtime
# Purpose: Complete Go toolchain from the official Go image
# Output: /srv/files/go<GO_VERSION>/ — full Go tree ready to COPY into base
# =============================================================================
FROM golang:${GO_VERSION}-alpine AS go-runtime
ARG GO_VERSION
RUN mkdir -p /srv/files/go${GO_VERSION} \
 && cp -r /usr/local/go/. /srv/files/go${GO_VERSION}/

# =============================================================================
# Stage: deno-runtime
# Purpose: Download the Deno binary from the official GitHub release
# Output: /srv/files/deno-<DENO_VERSION>/bin/deno
# Note: Uses debian:bookworm-slim directly (needs unzip, not in download-base)
# =============================================================================
FROM debian:bookworm-slim AS deno-runtime
ARG DENO_VERSION
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
        curl \
        ca-certificates \
        unzip \
 && rm -rf /var/lib/apt/lists/*
RUN mkdir -p /srv/files/deno-${DENO_VERSION}/bin \
 && curl -fL \
        -o /tmp/deno.zip \
        -- "https://github.com/denoland/deno/releases/download/v${DENO_VERSION}/deno-x86_64-unknown-linux-gnu.zip" \
 && cd /tmp \
 && unzip deno.zip \
 && mv /tmp/deno /srv/files/deno-${DENO_VERSION}/bin/deno \
 && chmod +x /srv/files/deno-${DENO_VERSION}/bin/deno

# =============================================================================
# Stage: task-binary
# Purpose: Download the go-task binary from the official GitHub release
# Output: /srv/files/task
# =============================================================================
FROM download-base AS task-binary
ARG TASK_VERSION
RUN curl -fL \
        -o /tmp/task.tar.gz \
        -- "https://github.com/go-task/task/releases/download/v${TASK_VERSION}/task_linux_amd64.tar.gz" \
 && tar -xzf /tmp/task.tar.gz -C /srv/files task \
 && rm /tmp/task.tar.gz \
 && chmod +x /srv/files/task

# =============================================================================
# Stage: node-runtime
# Purpose: Node.js with pnpm; openspec and opencode installed globally
# Output:
#   /srv/files/node/bin/node        — the node binary
#   /srv/files/node/lib/            — pnpm global node_modules
# Strategy: set npm global prefix to /srv/files/node so that
#   `pnpm add -g` places packages under /srv/files/node/lib/node_modules/
#   with a predictable, hash-free path.
# opencode-ai: npm wrapper that runs postinstall to download the ELF binary
#   into node_modules/opencode-ai/bin/opencode.exe — that ELF is the actual
#   executable; no node invocation needed for the wrapper.
# =============================================================================
FROM node:${NODE_VERSION}-bookworm-slim AS node-runtime
ARG PNPM_VERSION
ARG OPENSPEC_VERSION
ARG OPENCODE_VERSION
RUN corepack enable \
 && corepack prepare pnpm@${PNPM_VERSION} --activate
# Set npm/node prefix to our staging dir so node_modules path is predictable
RUN mkdir -p /srv/files/node/bin /srv/files/node/lib
# Copy node binary into the staging prefix
RUN cp /usr/local/bin/node /srv/files/node/bin/node
# Configure pnpm to use /srv/files/node as its global prefix
# global-dir sets where packages land; global-bin-dir for shims (not used)
RUN pnpm config set global-dir /srv/files/node/lib \
 && pnpm config set global-bin-dir /srv/files/node/bin
# Install openspec and opencode globally into /srv/files/node/lib/
# PATH includes node so postinstall scripts can run
ENV PATH="/usr/local/bin:${PATH}"
RUN pnpm add -g @fission-ai/openspec@${OPENSPEC_VERSION} \
 && pnpm add -g opencode-ai@${OPENCODE_VERSION}

# =============================================================================
# Stage: base
# Purpose: Assembly — all runtimes combined into the devcontainers base image
# =============================================================================
FROM mcr.microsoft.com/devcontainers/base:ubuntu-22.04 AS base
ARG GO_VERSION
ARG DENO_VERSION

# --- Go ---
COPY --from=go-runtime /srv/files/go${GO_VERSION} /opt/go${GO_VERSION}
RUN update-alternatives --install /usr/local/bin/go go /opt/go${GO_VERSION}/bin/go 100 \
        --slave /usr/local/bin/gofmt gofmt /opt/go${GO_VERSION}/bin/gofmt
ENV GOROOT=/opt/go${GO_VERSION} \
    GOPATH=/home/vscode/go

# --- Deno ---
COPY --from=deno-runtime /srv/files/deno-${DENO_VERSION} /opt/deno-${DENO_VERSION}
RUN update-alternatives --install /usr/local/bin/deno deno /opt/deno-${DENO_VERSION}/bin/deno 100

# --- Task ---
COPY --from=task-binary /srv/files/task /usr/local/bin/task

# --- Node / openspec / opencode ---
# Copy the node binary and the pnpm-global package tree
COPY --from=node-runtime /srv/files/node /opt/node
# Only node binary goes on PATH — pnpm, npm, npx intentionally absent
ENV PATH="/opt/node/bin:${PATH}"
# Create explicit wrapper scripts for the two permitted CLI tools.
# Paths are resolved at build time via find to handle pnpm's versioned
# internal directory structure (avoids hard-coding hash path).
# openspec: JS entrypoint — invoke via node
RUN OPENSPEC_JS=$(find /opt/node -name 'openspec.js' -path '*/@fission-ai/openspec/bin/*' | head -1) \
 && [ -n "$OPENSPEC_JS" ] || { echo "ERROR: openspec.js not found"; exit 1; } \
 && printf '#!/bin/sh\nexec /opt/node/bin/node %s "$@"\n' "$OPENSPEC_JS" \
        > /usr/local/bin/openspec \
 && chmod +x /usr/local/bin/openspec
# opencode: ELF binary placed by postinstall at bin/opencode.exe — invoke directly
RUN OPENCODE_EXE=$(find /opt/node -name 'opencode.exe' -path '*/opencode-ai/bin/*' | head -1) \
 && [ -n "$OPENCODE_EXE" ] || { echo "ERROR: opencode.exe not found"; exit 1; } \
 && printf '#!/bin/sh\nexec %s "$@"\n' "$OPENCODE_EXE" \
        > /usr/local/bin/opencode \
 && chmod +x /usr/local/bin/opencode
# Build-time guard: pnpm/npm/npx must not be reachable on PATH
RUN ! command -v pnpm 2>/dev/null \
 && ! command -v npm 2>/dev/null \
 && ! command -v npx 2>/dev/null

# =============================================================================
# Stage: ci
# Purpose: CI-appropriate environment layered on base
# =============================================================================
FROM base AS ci
ENV CI=true \
    DEBIAN_FRONTEND=noninteractive

# =============================================================================
# Stage: final
# Purpose: Developer environment layered on ci — SSH agent, XDG, build tracking
# =============================================================================
FROM ci AS final

# SSH agent profile script
COPY ssh-agent.sh /etc/profile.d/ssh-agent.sh
RUN echo "source /etc/profile.d/ssh-agent.sh" >> /home/vscode/.bashrc \
 && echo "source /etc/profile.d/ssh-agent.sh" >> /home/vscode/.zshrc

# VS Code server directories
RUN mkdir -p \
        /home/vscode/.vscode-server/extensions \
        /home/vscode/.vscode-server/data/User \
 && chown -R vscode:vscode /home/vscode/.vscode-server

# XDG environment variables and pager
ENV XDG_CONFIG_HOME=/home/vscode/.config \
    XDG_DATA_HOME=/home/vscode/.local/share \
    XDG_CACHE_HOME=/home/vscode/.cache \
    XDG_STATE_HOME=/home/vscode/.local/state \
    PAGER=cat

# Build tracking
RUN date +%s > /container_build_id
