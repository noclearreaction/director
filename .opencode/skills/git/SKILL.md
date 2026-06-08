---
name: git
description: Use for local Git workflow, including status, diffs, logs, branches, staging, commits, commit messages, commitlint, remotes, and PR preparation. This skill does not cover GitHub Issue creation.
compatibility: opencode
metadata:
  domain: version-control
  interface: cli
---

## What I do

I provide a safe, repeatable Git workflow for local repository work.

Use this skill whenever work involves:

- repository status
- diffs
- staged changes
- recent history
- branches
- remotes
- staging
- commits
- commit messages
- commitlint
- PR preparation

This skill covers local `git` command-line workflow. Use a separate GitHub Issue skill for GitHub Issues, labels, comments, and project-management state.

## Boundaries

Do not run destructive Git commands without explicit User approval.

Do not use:

- `git reset --hard`
- `git checkout --`
- force-push
- commit amend
- hook skipping
- Git config mutation

Do not commit, push, or create PRs unless explicitly requested.

Do not stage unrelated files.

Do not infer completion from Git state alone.

If a requested action requires a command outside the active agent's permissions, stop and ask the User how to proceed.

## Tool policy

Use local `git` CLI for local repository inspection and local branch, diff, and history work.

Use `gh` for GitHub remote state when available and permitted.

Do not silently substitute tools across these boundaries. If the expected interface is unavailable, blocked by permission, or fails, state the limitation and ask how to proceed.

## Standard inspection

Before staging, committing, pushing, or preparing a PR, inspect local state.

Use:

```bash
git status --short
git diff
git diff --staged
git log --oneline -10
```

## Conventional Commits and Branch Rules

To maintain high repository hygiene and strict alignment with Symphony's Constitution, you MUST adhere to the following rules:

1. **Zero Commits to Main**: You are strictly prohibited from making direct commits to the `main` branch. All work must be done on dedicated topic branches named `change/<name>`.
2. **Single-Topic Branching**: Each branch must correspond to exactly one OpenSpec change.
3. **Conventional Commit Standard**: Every commit message you write MUST conform to Conventional Commits:
   `<type>(<scope>): <subject>`
   Allowed types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`.
4. **Atomic, Logical Commits**: Commit on each logical unit of work (e.g. after completing a specific file, class, or task in `tasks.md`). Never combine multiple unrelated changes into a single mega-commit.
5. **Validation Gating**: Run the linter command to validate your commit message format before committing, or ensure that the git hooks are not bypassed:
   ```bash
   deno run --allow-read --allow-env bin/commit-lint.ts "your commit message"
   ```