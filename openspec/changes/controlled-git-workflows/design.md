## Context

Currently, the Director repository does not have active validation for commit messages or rigid branch-level gating. Both human operators and AI agents could accidentally commit directly to `main` or make uncoordinated "mega-commits" containing multiple unrelated changes. To align with Symphony's Constitution (Transparency & Intent Traceability), we need to formalize branch lifecycles, enforce Conventional Commits, implement a commit linter, and make AI agents explicitly aware of these constraints.

## Goals / Non-Goals

**Goals:**
- **Zero Commits to Main**: Block direct commits to `main`. All work must proceed on single-topic feature branches.
- **Single-Topic Branches**: Enforce that each branch (of the form `change/<change-name>`) is bound to a single OpenSpec change.
- **Atomic Commits**: Commits must be made on each logical unit of work (e.g., specific tasks, refactors, or files) rather than delayed until the end of the feature.
- **Deno-based Commit Linter**: Design a zero-dependency Deno TypeScript script `bin/commit-lint.ts` that validates commit messages against Conventional Commits.
- **Agent Alignment**: Update AI agent system instructions (`builder.md`, `orchestrator.md`, `designer.md`) to be fully aware of and enforce the branch and commit constraints.

**Non-Goals:**
- **NPM Toolchain Installation**: Avoid installing complex npm dependencies or heavy packages (like husky or commitlint npm package) in the root. We prefer a native, fast Deno validation approach.
- **Automatic Merge Resolution**: Merging and conflict resolution remain a human-in-the-loop task.

## Decisions

### Decision 1: Zero-Dependency Deno Commit Linter
- **Choice**: Implement a lightweight TypeScript script `bin/commit-lint.ts` that runs with Deno.
- **Rationale**: Keeps the codebase minimal, leveraging the existing Deno setup used for `bin/director-start.ts` and `bin/director-submit.ts`. It parses the commit message against conventional commit patterns: `^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(?:\(.+\))?: .{1,100}$`.

### Decision 2: Pre-commit Gating in Submit Script
- **Choice**: Integrate commit lint validation in `bin/director-submit.ts`.
- **Rationale**: Before automatically committing the synchronized specification updates, the submit script must ensure the message complies with Conventional Commits. We can also add instructions for installing a local git `.git/hooks/commit-msg` or `pre-commit` hook that invokes `bin/commit-lint.ts`.

### Decision 3: Enforcing Logical Commit Units
- **Choice**: Specify rules for Builder agents to commit frequently—specifically, upon completing individual tasks in `tasks.md` or separate files.
- **Rationale**: Prevents monolithic commits, making history easy to read, revert, and audit.

## Risks / Trade-offs

- **[Risk] Agent bypasses linter**: An agent could execute direct `git commit` without running the linter if using general bash access.
  - **Mitigation**: Tighten builder bash permissions to only allow specific git commands or require asking (which they already do), and clearly specify the commit lint requirement in their system prompt.
- **[Risk] Restrictive branch conventions block urgent fixes**: Prohibiting direct commits to `main` might feel heavy for trivial chores.
  - **Mitigation**: All chores must follow the same flow (create a change branch, perform the chore, sync and PR). This guarantees total traceability.
