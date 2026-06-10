# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

- Run the app locally: `go run .`
  - Defaults to `PORT=18080` and `DB_PATH=./data/app.db`.
  - Visit `http://localhost:18080`.
- Run all tests: `go test ./...`
- Run one package's tests: `go test ./render` or `go test ./storage`
- Run one test by name: `go test ./storage -run TestName`
- Run race-enabled tests: `go test -race ./...`
- Run coverage: `go test -cover ./...`
- Static checks: `go vet ./...`
- Format Go files: `gofmt -w path/to/file.go`
- Build binary: `go build -o mdshare .`
- Build and run development container: `docker-compose -f docker-compose.dev.yml up -d --build`
- Run published image: `docker-compose up -d`

## Runtime Configuration

- `PORT`: HTTP port, default `18080`.
- `DB_DRIVER`: database backend selector, default `sqlite`. Supported values: `sqlite`, `libsql`.
- `DB_PATH`: SQLite database path or SQLite DSN, default `./data/app.db`; `main.go` creates the parent directory for file-backed SQLite paths.
- `DB_URL`: libSQL/Turso database URL, required when `DB_DRIVER=libsql`.
- `DB_AUTH_TOKEN`: libSQL/Turso auth token, optional unless the remote database requires it.
- Docker Compose mounts `./data` to persist local SQLite data outside the container.
- Vercel defaults to `DB_DRIVER=sqlite` with `DB_PATH=/tmp/comelymd/comelymd.db`; this is ephemeral scratch storage, not durable persistence.

## Architecture

ComelyMD is a small Go 1.21 Markdown sharing service built on `net/http` with selectable SQLite or libSQL persistence.

- `main.go` wires runtime configuration, resolves the selected database backend, initializes persistence through `storage.InitDB`, then starts the HTTP server through `handler.Run`.
- `handler/server.go` owns route registration and server timeouts. Routes are intentionally minimal: `/` for the editor page, `/api/pages` for page creation, `/p/` for shared-page viewing, and `/static/` for assets.
- `handler/api.go` contains request handling for the product flow:
  - Parses form or multipart submissions with a 5 MB body limit.
  - Converts Markdown to sanitized HTML before persistence.
  - Supports burn-after-read, optional generated passwords, and fixed expiration windows.
  - Renders `templates/*.html` through the package-level parsed template set.
- `render/markdown.go` is the Markdown security boundary. It uses Goldmark with GFM and unsafe HTML enabled, then sanitizes with Bluemonday. Math spans/blocks are temporarily replaced before Markdown conversion and restored after sanitization as escaped text so KaTeX-style syntax survives without becoming executable HTML.
- `storage/database.go` is the persistence layer. It keeps a package-level `*sql.DB`, opens either local SQLite or libSQL based on runtime configuration, creates or migrates the `pages` table on startup, periodically deletes expired rows, and enforces expiration again during reads. `storage/idgen.go` generates cryptographically random Base62 IDs and passwords.
- Frontend assets are server-rendered templates plus static files. Client-side Markdown enhancements, theme behavior, code highlighting, KaTeX/Mermaid display, and page actions live under `templates/` and `static/` rather than a JS build pipeline.
- `userscript/` is a separate browser userscript integration that posts Markdown into `/api/pages`; keep API compatibility in mind when changing creation semantics or CORS behavior.

## Pull Request Workflow

### Branch Rules

- Treat `main` as the stable branch.
- Do not develop directly on `main`.
- Create a focused branch for each change, using names such as `feat/...`, `fix/...`, `docs/...`, `chore/...`, or `release/...`.
- Do not push directly to `main` for normal feature, bug fix, documentation, or release work.
- Open a pull request for every branch intended to enter `main`.

### Before Pushing

Before pushing a branch to GitHub:

1. Confirm working tree state:

   ```bash
   git status --short
   ```

2. Review staged and unstaged changes:

   ```bash
   git diff
   git diff --cached
   ```

3. Run the baseline validation:

   ```bash
   go test ./...
   go vet ./...
   ```

4. For Markdown rendering, storage, handler, API, CORS, password, expiration, or burn-after-read changes, add targeted tests or manually verify the affected flow.
5. Pull latest remote state before pushing when working from an older branch. Do not chain pull and push in one command where push can run after a failed pull.

### PR Title

Use the PR title format:

```text
comelyMD-#<issue-number>: <english summary>
```

For PRs without a related issue, use:

```text
MINOR: <english summary>
```

Examples:

```text
comelyMD-#12: preserve math blocks during Markdown sanitization
MINOR: add Claude Code project guidance
```

### PR Body

Every PR should include:

````md
## Summary

-
-

## Type

- [ ] Feature
- [ ] Bug fix
- [ ] UI/UX
- [ ] API behavior
- [ ] Security/sanitization
- [ ] Database/storage
- [ ] Release/build
- [ ] Documentation

## Verification

Commands run:

```text

```

Manual cases tested:

-

## Areas Affected

- [ ] Markdown rendering / sanitization
- [ ] Page creation API
- [ ] Password-protected pages
- [ ] Burn-after-read behavior
- [ ] Expiration cleanup
- [ ] SQLite persistence
- [ ] Templates / static frontend
- [ ] Userscript compatibility
- [ ] Docker / deployment
- [ ] Not applicable

## Documentation Check

- [ ] Updated README / CLAUDE.md / userscript docs if behavior changed
- [ ] Confirmed docs update is not needed
- [ ] Checked docs for local paths, private logs, real user content, and secrets

## Risk And Follow-Up

-
````

- Run the local validation baseline before opening or updating a PR; current GitHub automation may not validate pull requests before merge.
- Merge only after review comments are resolved and required local validation is recorded in the PR.
- Prefer squash merge for normal feature, fix, docs, and chore PRs.
- Release PRs may use a merge commit if preserving release context is useful.

### Existing PR Branch Rule

When working on a branch that already has an open PR:

1. Identify the PR associated with the current branch.
2. Check review decision, unresolved review threads, comments, and CI status.
3. Treat blocking review feedback as part of the current task.
4. Resolve feedback before adding unrelated new work.
5. Rerun validation and update the PR with what changed.

## Security-Sensitive Areas

- `render.MarkdownToHTML` controls untrusted Markdown/HTML sanitization; changes here can create XSS risk.
- `handler.CreatePageHandler` accepts cross-origin form posts for the userscript integration and currently sets permissive CORS for `/api/pages`.
- Shared page HTML is rendered as `template.HTML` only after server-side sanitization; do not bypass the render pipeline for user content.
- Database access uses parameterized SQL; keep it that way for IDs, passwords, and page content.

## Project Notes

- The module name is `mdshare`, so internal imports use paths like `mdshare/handler`.
- The project has no test files at the time this document was created; `go test ./...` should still be used as the baseline regression check.
- The Dockerfile is a multi-stage build using `CGO_ENABLED=0` with `modernc.org/sqlite`, so it does not require CGO-enabled SQLite bindings.
