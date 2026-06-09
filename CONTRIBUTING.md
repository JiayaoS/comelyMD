# Contributing to ComelyMD

Thanks for contributing to ComelyMD. This project is a small Go Markdown sharing service with SQLite persistence, server-rendered templates, static frontend assets, and a browser userscript integration.

## Development Setup

Requirements:

- Go 1.21+
- Docker and Docker Compose, if validating container behavior

Run locally:

```bash
go run .
```

By default the app listens on `http://localhost:18080` and stores SQLite data at `./data/app.db`.

Run with Docker for development:

```bash
docker-compose -f docker-compose.dev.yml up -d --build
```

## Validation

Before opening or updating a pull request, run the baseline checks:

```bash
go test ./...
go vet ./...
```

Useful targeted commands:

```bash
go test ./render
go test ./storage
go test ./handler -run TestName
go test -race ./...
go test -cover ./...
go build -o mdshare .
```

Format edited Go files with:

```bash
gofmt -w path/to/file.go
```

For Docker or deployment changes, also validate the affected target:

```bash
docker-compose -f docker-compose.dev.yml up -d --build
```

For Vercel configuration changes, verify `vercel.json` still points SQLite at `/tmp` and document any persistence limitations in the PR.

## Areas To Test Manually

Manual verification is expected when a change affects behavior that tests do not cover yet.

- Markdown rendering or sanitization: submit Markdown with headings, tables, code blocks, raw HTML, math syntax, and Mermaid/KaTeX-related content.
- Page creation API: verify `/api/pages` with form or multipart submissions.
- Password protection: create a password-protected page, test wrong and correct passwords.
- Burn-after-read: open the shared page once, then confirm the second visit no longer renders it.
- Expiration: test the affected expiration option or storage cleanup path.
- Templates/static frontend: verify the home page, shared page, mobile layout, copy/download/export actions when touched.
- Userscript compatibility: preserve the `/api/pages` contract and CORS behavior unless the change intentionally updates the integration.

## Branch And PR Workflow

- Treat `main` as the stable branch.
- Do not develop directly on `main`.
- Create focused branches such as `feat/...`, `fix/...`, `docs/...`, `chore/...`, or `release/...`.
- Do not push directly to `main` for normal feature, bug fix, documentation, or release work.
- Open a pull request for every branch intended to enter `main`.

Before pushing:

```bash
git status --short
git diff
git diff --cached
go test ./...
go vet ./...
```

When working from an older branch, update from the latest remote state before pushing. Do not chain pull and push in one command where push can run after a failed pull.

Run the local validation baseline before opening or updating a PR; current GitHub automation may not validate pull requests before merge.

## Commit Messages

Use concise conventional-style commit subjects:

```text
<type>: <short summary>
```

Recommended types:

- `feat`
- `fix`
- `docs`
- `refactor`
- `test`
- `chore`
- `ci`
- `perf`

For user-facing, security-sensitive, storage, API, or deployment changes, include a short body explaining the behavior change, implementation note, compatibility impact, and verification.

## Pull Request Template

Use the PR title format:

```text
comelyMD-#<issue-number>: <english summary>
```

For PRs without a related issue, use:

```text
MINOR: <english summary>
```

PR body:

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

## Security-Sensitive Changes

Be especially careful with:

- `render.MarkdownToHTML`, because it controls untrusted Markdown and HTML sanitization.
- `handler.CreatePageHandler`, because it accepts user input and cross-origin userscript requests.
- Shared page rendering with `template.HTML`, which should only receive sanitized content.
- SQLite queries in `storage/`, which should stay parameterized.

For these areas, include targeted tests or clear manual verification in the PR.
