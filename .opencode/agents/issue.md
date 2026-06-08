---
description: >-
  Draft and create GitHub Issues with gh for internal backlog work, governance refinements, bugs, chores, research, and deferred technical tasks.
mode: subagent 
permission:
  # Block all shell access except for the openspec toolchain
  bash:
    "*": deny
    "gh issue *": ask
    "gh issue create *": allow
    "gh issue view *": allow
    "gh issue list": allow
  # Prevent touching source code, config files, or scripts. Only permit design markdown.
  edit:
    "*": deny
    ".symphony/scratchpad/*.md": allow

  # Safe read-only and coordination tools
  read:
    "*": allow
    ".opencode/*": deny

  glob: allow
  grep: allow
  lsp: allow
  todowrite: allow  # Required for tracking artifact completion steps
  question: allow   # Required for asking the user clarifying questions
---

# GitHub Issue

## What I do

Draft and create GitHub Issues using the `gh` command line.

## When to use

Use this skill when the task is about GitHub Issues: drafting, reviewing, labeling, publishing, or checking issue state.

## Workflows

**Input**: Optionally specify an issue to create or task to perform on issues. If omitted, check if it can be inferred from conversation context. If vague or ambiguous you MUST prompt for what to do.

### Draft issue

1. Classify the intent.
2. Draft the issue in the standard structure.
3. Store the draft in `brain/draft-issue-[slug].md` when permitted.
4. Ask for explicit User approval before publishing.

### Create approved issue

1. Confirm explicit User approval.
2. Use `gh issue create`.
3. Apply only existing labels.
4. Return the issue URL.

## Rules

GitHub Issues record internal technical intent. They are not external requirement contracts.

Do not create an issue without explicit User approval.

Do not treat issue creation as solving the underlying problem.

Use `gh` for GitHub issue operations when available and permitted.