# CephTower Claude Code Guide

## Development Direction

This project moves quickly and does not preserve backward compatibility unless the user
explicitly requests it.

- When requirements change, update the implementation fully to match the new direction.
- Do not add compatibility layers, legacy fallbacks, dual API paths, old configuration
  aliases, database compatibility shims, or transitional code for previous behavior.
- Remove or replace obsolete database, configuration, API, backend, frontend, and test
  logic that conflicts with the new requirement.
- Prefer a clean current implementation over preserving old behavior for hypothetical
  existing consumers.
- If preserving compatibility appears necessary, stop and confirm it with the user before
  adding compatibility logic.

Claude Code must follow the shared project commit convention in
[docs/commit-convention.md](docs/commit-convention.md).

Key constraints:

- Do not create a Git commit unless the user explicitly asks for one.
- Before committing, inspect `git status --short` and stage only files relevant to the
  requested change.
- Never revert, discard, or overwrite unrelated user changes.
- Prefer small, atomic commits with one clear purpose.
- Run relevant checks before committing:
  - Backend changes: `make test-backend`
  - Frontend changes: `make test-frontend` after dependencies are installed
  - Documentation-only changes: no build is required
- Use the commit message format `type: summary`, followed by a blank line and body
  bullets.
- Write all commit titles and body details in English.
- Start each body bullet with `- ` and keep each body line at or below 90 characters.
- Start the description after `- ` with a lowercase letter.
- For wrapped bullet text, indent continuation lines by two spaces to align with the text.
- When creating a GitHub MR or PR, use `.github/PULL_REQUEST_TEMPLATE.md`.
- Keep the PR content in English.
- Fill every relevant PR template section and remove items that do not apply.

Examples:

```text
docs: move multilingual readmes under docs

- move translated README files into the docs/readme directory
- update root README links so each language points to the new location
```
