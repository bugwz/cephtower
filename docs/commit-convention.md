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
[TYPE]: Summary

- Change detail
- Another change detail
```

Rules:

- Write all commit content in English.
- The title must be `[TYPE]: Summary`.
- `TYPE` must be uppercase and enclosed in square brackets.
- The summary after `]: ` must start with an uppercase letter.
- Leave exactly one blank line between the title and the body.
- The body is required.
- Each body bullet must start with `- `.
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
| `FEAT` | User-facing feature or new capability |
| `FIX` | Bug fix |
| `DOCS` | Documentation-only change |
| `STYLE` | Formatting or visual style change without behavior change |
| `REFACTOR` | Code restructuring without behavior change |
| `TEST` | Test additions or test-only changes |
| `CHORE` | Repository maintenance |
| `BUILD` | Build system or dependency change |
| `CI` | CI workflow change |
| `PERF` | Performance improvement |
| `REVERT` | Revert a previous commit |

## 5. Examples

```text
[CHORE]: Initialize project structure

- Add Go backend service skeleton with configuration and HTTP API packages
- Add Vue frontend scaffold with Vite, TypeScript, and initial cluster overview UI
- Add MIT license, README, Makefile, and repository ignore rules
```

```text
[DOCS]: Update agent commit convention

- Add shared commit message policy for Codex and Claude Code project entry files
- Require uppercase commit types, English summaries, and bullet-based commit bodies
- Document the 90-character line limit for each body bullet
```

Wrapped body bullet example:

```text
[DOCS]: Clarify commit body wrapping

- Document how long commit body bullets should wrap when they exceed the line length
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
