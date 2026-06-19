# How to work on cts (safely)

Read this before running commands or touching the code. This tool **deletes real things** — the discipline here is not optional.

## Commands you CAN use (dev)

| Command | What it does |
|---|---|
| `go run . scan` | Runs the scan (read-only, safe) |
| `go test ./...` | Runs the tests |
| `go test -race ./...` | Tests with the race detector |
| `go vet ./...` | Compiler static analysis |
| `golangci-lint run ./...` | Lint |
| `go build -o cts.exe .` | Builds the binary |
| `./scripts/check.sh` | Full gate: fmt + vet + lint + test + build |
| `gofmt -w .` | Formats |

## Commands you can NOT (dangerous)

- ❌ **Running destructive `cts cut`/`cts purge` against the real machine during dev/testing.** They remove the user's files. In dev, only with dry-run, or against a test directory.
- ❌ **Testing against real paths** (`~/.claude`, `~/.agents`, `~/.codex`...). Every test uses `t.TempDir()`. A test that deletes must not touch the real home.
- ❌ **Committing `cts.exe`** or any binary (it's in `.gitignore`).
- ❌ **Committing with a red gate.** Test/lint/build must pass first.
- ❌ **`git push --force`, destructive rebase, reset --hard** without a clear reason.

## The tool's safety model

1. **Dry-run is the default.** Without an explicit flag, cts only shows — it does not remove.
2. **Confirmation** before any real removal.
3. **Backup** in `.cts-backups/` before deleting.
4. **Check the target** before removing. If the content contradicts what was described, stop and warn.

## Workflow (XP loop)

1. **Plan** — smallest slice that delivers value. One scanner/feature at a time.
2. **Test** — write the table-driven test first (`t.TempDir()`), confirm the behavior.
3. **Implement** — minimal code to pass.
4. **Refactor** — clean up with the test green.
5. **Gate** — `./scripts/check.sh` green.
6. **Commit** — Conventional Commits, no agent co-author. Pause for review.

## Git

- Branch per feature (`feat/agents-scanner`), never directly on `main`.
- Conventional Commits: `feat:`, `fix:`, `refactor:`, `docs:`, `chore:`, `test:`.
- Subject describes what was done; body only when the "why" isn't obvious.
- No agent `Co-Authored-By`.

## Code conventions

Idiomatic Go — see `~/.claude/go-conventions.md`. Summary: no inheritance (composition), small interface defined at the consumer, errors as values with `%w`, small package per capability, no generic `util`, small and cohesive files.
