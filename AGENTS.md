# CephTower Agent Guide

## Commit Discipline

Codex and other coding agents working in this repository must follow the project commit
convention in [docs/commit-convention.md](docs/commit-convention.md).

Claude Code users should also see [CLAUDE.md](CLAUDE.md). Both files point to the same
shared convention document.

Key constraints:

- Do not create a Git commit unless the user explicitly asks for one.
- Before committing, inspect `git status --short` and include only files relevant to the
  requested change.
- Never revert, discard, or overwrite unrelated user changes.
- Prefer small, atomic commits with one clear purpose.
- Run relevant checks before committing. For backend changes, run `make backend-test`.
  For frontend changes, run `make frontend-build` after dependencies are installed.
- If a check cannot be run, mention the reason in the final response.
- Use the commit message format `type: summary`, followed by a blank line and body
  bullets.
- Write all commit titles and body details in English.
- Start each body bullet with `- ` and keep each body line at or below 90 characters.
- Start the description after `- ` with a lowercase letter.
- For wrapped bullet text, indent continuation lines by two spaces to align with the text.

## Pull Request Discipline

- When creating a GitHub MR or PR, use `.github/PULL_REQUEST_TEMPLATE.md`.
- Keep the PR content in English.
- Fill every relevant section and remove checklist items that do not apply.

Examples:

```text
docs: move multilingual readmes under docs

- move translated README files into the docs/readme directory
- update root README links so each language points to the new location
```
