## Why

Currently, there is a risk of uncoordinated commits, direct modifications to the `main` branch, lack of structured commit messages, and a lack of clear protocols for human-agent co-authoring. Defining rigid Git rules and specifying a multi-agent "Core Loop" methodology ensures strict alignment with Symphony's Constitution (Principle 2: Transparency, Principle 3: Intent Traceability), enforcing that all changes are tracked, auditable, and trace-linked to explicit specifications.

## What Changes

- **Strict Git Rules**: Establish a strict workflow where all development occurs on single-topic branches (no direct commits to `main`), commits are made on each logical unit of work, and all commits conform to Conventional Commits validated via commitlint.
- **Multi-Agent "Core Loop" Specification**: Define the precise lifecycle from GitHub Issue to OpenSpec Change to Implementation Tasks to pull requests, detailing handoffs between Orchestrator, Designer, Issue, and Builder roles.
- **Agent Awareness**: Ensure AI agents (specifically the General Orchestrator) are fully aware of and conform to these branch, commit, and core loop protocols.
- **Verification Gates**: Introduce automated pre-commit linting or guidance to verify commit message compliance.

## Capabilities

### New Capabilities
- `controlled-git-workflows`: Establishes the rules, specifications, and protocols for managing git branches, commit conventions, and coordination loops.

### Modified Capabilities
- `director-workflow`: Updates the Director workflow and agent guidelines to align with the new controlled git workflow and multi-agent core loop execution pattern.

## Impact

- **Developer/Agent Environment**: Direct commits to `main` are prohibited. Branch creation, naming, staging, and commit procedures are tightly constrained.
- **Agent Configuration**: Orchestrator, Designer, Issue, and Builder agent definitions are updated or clarified to ensure they enforce and operate within this flow.
- **Commit Linting**: Install and configure commitlint and husky/pre-commit hooks (if applicable) to validate conventional commit messages.
