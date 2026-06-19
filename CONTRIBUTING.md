# Contributing to cts

Thanks for your interest in improving `cts`. This is a small, focused Go CLI — contributions that keep it simple, safe and well-tested are very welcome.

## Ground rules (read first)

`cts` **deletes files from the user's machine.** Safety is not optional:

- **Tests must use `t.TempDir()`** — never a real user path (`~/.claude`, `~/.config`, `AppData`, ...). A test that deletes must never touch a real home directory.
- **Never** run a destructive `cts purge --yes` / `cts clean` against your real machine while developing. Use dry-run, or point it at a throwaway directory.
- Removal stays **dry-run by default**, always **confirms**, and always **backs up** before deleting. Don't weaken that.

## Prerequisites

- Go (the version in [`go.mod`](go.mod))
- `golangci-lint` (optional locally; CI runs it)

## Build & test

```bash
go build -o cts .          # build the binary
go test ./...              # run the tests
./scripts/check.sh         # full local gate: fmt + vet + lint + race tests + build
```

Open a PR only when `./scripts/check.sh` is green.

## Workflow

`cts` is built in small, reviewable increments:

1. **Plan** — the smallest slice that delivers value (usually one scanner or one behavior).
2. **Test first** — write a table-driven test with `t.TempDir()` that describes the behavior.
3. **Implement** — the minimal code to pass.
4. **Refactor** — clean up while the tests are green.
5. **Gate** — `./scripts/check.sh` passes.

## Code conventions

Idiomatic Go (see [`docs/WORKING.md`](docs/WORKING.md)):

- Composition over inheritance; small interfaces defined at the consumer.
- Errors as values, wrapped with `%w` and context.
- Small packages named by capability — no generic `util`.
- Keep dependencies **injected** (PATH lookups, command execution, config roots) so the core stays testable.

The architecture and the "how to add a scanner" recipe live in [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md). Design decisions are recorded in [`docs/adr/`](docs/adr/).

## Commits & pull requests

- **Conventional Commits**: `feat:`, `fix:`, `refactor:`, `docs:`, `chore:`, `test:`.
- One logical change per commit; subject describes what was done.
- Branch per feature (`feat/...`), never commit directly to `main`.
- Keep PRs focused and reviewable. Explain the "why" when it isn't obvious.

## Adding a new scanner

See **How to add a new scanner** in [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md). In short: implement the small `Scanner` interface in `internal/scan/<category>`, hide the "is it dead?" rule behind a private function, test it in isolation with `t.TempDir()`, and register it in `main`.
