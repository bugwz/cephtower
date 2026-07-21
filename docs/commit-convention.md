# Commit Convention

This document defines how agents should create commits for CephTower.

## 1. Agent Entry Points

The shared commit policy lives in this file. Tool-specific entry files should stay short
and point here:

| Tool | Entry Point |
|------|-------------|
| Codex | `AGENTS.md` |
| Codex repo skill | `.agents/skills/commit-convention/SKILL.md` |
| Claude Code | `CLAUDE.md` |
| GitHub PR | `.github/PULL_REQUEST_TEMPLATE.md` |

Do not use symlinks for these entry points. Some editors, hosted agents, and sandboxed
runners resolve symlinks differently, so plain markdown files are more portable.

## 2. When To Commit

Only create a commit when the user explicitly requests it.

Before committing:

1. Run `git status --short`.
2. Review the diff for all files that will be staged.
3. Stage only files related to the current request.
4. Run the relevant checks.
5. Commit with the project message format.

## 3. Message Format

Use this format for every commit:

```text
type: summary

- change detail
- another change detail
```

Rules:

- Write all commit content in English.
- The title must be `type: summary`.
- `type` must be lowercase and must not be enclosed in square brackets.
- Separate the type and summary with `: `.
- The summary after `: ` must start with a lowercase letter.
- Leave exactly one blank line between the title and the body.
- The body is required.
- Each body bullet must start with `- `.
- The description after `- ` must start with a lowercase letter.
- Each body bullet should describe a concrete change detail.
- Do not insert blank lines between body bullets.
- Keep each body line at or below 90 characters.
- If a bullet exceeds 90 characters, wrap the continuation line with two leading spaces.
- Continuation lines must align with the first character after `- ` in the original
  bullet.
- Do not end the title with a period.
- When using `git commit`, pass the body as one paragraph or use a message file; do
  not pass each bullet with a separate `-m`, because Git inserts blank lines between
  separate message paragraphs.

## 4. Types

| Type | Meaning |
|------|---------|
| `feat` | User-facing feature or new capability |
| `fix` | Bug fix |
| `docs` | Documentation-only change |
| `style` | Formatting or visual style change without behavior change |
| `refactor` | Code restructuring without behavior change |
| `test` | Test additions or test-only changes |
| `chore` | Repository maintenance |
| `build` | Build system or dependency change |
| `ci` | CI workflow change |
| `perf` | Performance improvement |
| `revert` | Revert a previous commit |

## 5. Examples

```text
chore: initialize project structure

- add Go backend service skeleton with configuration and HTTP API packages
- add React frontend scaffold with Ant Design, Vite, and TypeScript
- add MIT license, README, Makefile, and repository ignore rules
```

```text
docs: update agent commit convention

- add shared commit message policy for Codex and Claude Code project entry files
- require lowercase commit types, English summaries, and bullet-based commit bodies
- document the 90-character line limit for each body bullet
```

Wrapped body bullet example:

```text
docs: clarify commit body wrapping

- document how long commit body bullets should wrap when they exceed the line length
  limit enforced by the project commit convention
```

## 6. Agent Safety Rules

- Never commit ignored reference material from `docs/references/`.
- Never include secrets from local configuration files.
- Never stage unrelated files just because they are present in the worktree.
- Do not amend, rebase, reset, or force-push unless the user explicitly requests it.
- If generated files are required, explain why they are included.

## 7. Pull Request Rules

- When creating a GitHub MR or PR, use `.github/PULL_REQUEST_TEMPLATE.md`.
- Write all PR content in English.
- Fill every relevant section.
- Remove checklist items or placeholders that do not apply.
