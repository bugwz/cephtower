---
name: commit-convention
description: Use when creating or preparing Git commits in the CephTower repository,
  reviewing staged changes, writing commit messages, or explaining commit policy for
  agents.
---

# CephTower Commit Convention

When the user asks to create or prepare a commit in this repository, follow these rules.

## Workflow

1. Confirm the user explicitly requested a commit.
2. Run `git status --short`.
3. Review the diff for files that may be staged.
4. Stage only files related to the user's request.
5. Run relevant checks:
   - Backend changes: `make backend-test`
   - Frontend changes: `make frontend-build` when dependencies are installed
   - Documentation-only changes: no build is required
6. Commit with the required message format.

For the complete policy, read
[docs/commit-convention.md](../../../docs/commit-convention.md).

## Message Format

Use:

```text
type: summary

- change detail
- another change detail
```

Rules:

- Write all commit content in English.
- `type` must be lowercase and must not be enclosed in square brackets.
- Separate the type and summary with `: `.
- The summary after `: ` must start with a lowercase letter.
- Leave exactly one blank line between the title and body.
- The body is required.
- Every body bullet must start with `- `.
- The description after `- ` must start with a lowercase letter.
- Do not insert blank lines between body bullets.
- Keep each body line at or below 90 characters.
- Wrap longer bullet text with continuation lines indented by two spaces.
- Continuation lines must align with the first character after `- ` in the original
  bullet.
- Do not end the title with a period.
- When using `git commit`, pass the body as one paragraph or use a message file; do
  not pass each bullet with a separate `-m`, because Git inserts blank lines between
  separate message paragraphs.

Allowed types:

`feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `build`, `ci`,
`perf`, `revert`.

Examples:

```text
docs: move multilingual readmes under docs

- move translated README files into the docs/readme directory
- update root README links so each language points to the new location
```

```text
docs: clarify commit body wrapping

- document how long commit body bullets should wrap when they exceed the line length
  limit enforced by the project commit convention
```

## Safety

- Do not commit unless explicitly asked.
- Do not stage unrelated user changes.
- Do not commit `docs/references/`, `.env`, build caches, or dependency directories.
- Do not amend, rebase, reset, or force-push unless explicitly requested.
- If checks cannot be run, report the reason clearly.

## Pull Requests

- When creating a GitHub MR or PR, use `.github/PULL_REQUEST_TEMPLATE.md`.
- Keep the PR content in English.
- Fill every relevant section and remove checklist items that do not apply.
