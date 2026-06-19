# cts — agent instructions

A Go CLI that removes dead skills/agents/plugins/MCP servers from the user's machine.
**This tool DELETES the user's stuff — safety comes before any feature.**

## Safety rule (non-negotiable)

- **Dry-run is the default.** Nothing is removed without an explicit execution flag + user confirmation.
- **Always back up before removing.** It goes to `.cts-backups/`.
- **NEVER** run destructive `cts cut`/`cts purge` against the real machine during dev or testing.
- **Tests use a temp directory** (`t.TempDir()`), never real user paths (`~/.claude`, etc.).
- Before removing anything: check the target. If what is there contradicts what was described, stop and warn.

## How to work here (pull on demand — don't bloat the context)

Read the file only when the case applies:

- **Commands (what you can / can NOT do) + safe workflow** → `docs/WORKING.md`
- **Architecture, module design, how to add a scanner** → `docs/ARCHITECTURE.md`
- **Design decisions (ADRs)** → `docs/adr/`
- **Idiomatic Go conventions** → `~/.claude/go-conventions.md`
- **Rafael's working rules (values, XP loop, principles)** → `~/.agents/AGENTS.md`

## Quick

- Full gate (fmt, vet, lint, test, build): `./scripts/check.sh`
- Run: `go run . scan`
- Test: `go test ./...`

Commit only with a green gate. Conventional commits. No agent co-author. Feature on a branch, never directly on main.
