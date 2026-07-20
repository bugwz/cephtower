# CephTower Claude Code Guide

Claude Code must follow the shared project commit convention in
[docs/commit-convention.md](docs/commit-convention.md).

Key constraints:

- Do not create a Git commit unless the user explicitly asks for one.
- Before committing, inspect `git status --short` and stage only files relevant to the
  requested change.
- Never revert, discard, or overwrite unrelated user changes.
- Prefer small, atomic commits with one clear purpose.
- Run relevant checks before committing:
  - Backend changes: `make backend-test`
  - Frontend changes: `make frontend-build` after dependencies are installed
  - Documentation-only changes: no build is required
- Use the commit message format `[TYPE]: Summary`, followed by a blank line and body
  bullets.
- Write all commit titles and body details in English.
- Start each body bullet with `- ` and keep each body line at or below 90 characters.
- For wrapped bullet text, indent continuation lines by two spaces to align with the text.
- When creating a GitHub MR or PR, use `.github/PULL_REQUEST_TEMPLATE.md`.
- Keep the PR content in English.
- Fill every relevant PR template section and remove items that do not apply.

Examples:

```text
[DOCS]: Move multilingual readmes under docs

- Move translated README files into the docs/readme directory
- Update root README links so each language points to the new location
```
