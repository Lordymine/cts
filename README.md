# cts — Cut The Shit

A Go CLI that finds and removes **dead skills, agents, plugins and MCP servers** from your machine — with dry-run, confirmation and backup. It grew out of a manual cleanup that turned into a tool.

## Why

AI tools (Claude Code, codex, opencode, pi, and various agents) leave junk scattered around: orphan skills, agent binaries you no longer use, plugin caches, orphan marketplaces, MCP servers configured and forgotten. This eats disk space and — worse — bloats the context injected into every prompt. `cts` automates the cleanup safely.

## Install

```bash
go build -o cts.exe .
```

## Usage

```bash
cts             # interactive mode: lists items, you pick what to remove, confirm (with backup)
cts scan        # read-only report: what is dead and how much space it uses
cts purge       # shows what it would remove (dead items only) — dry-run
cts purge --yes # actually removes the dead items, with a backup in .cts-backups/
```

## Status

MVP under construction, increment by increment:

- [x] `scan` for skills
- [x] `scan` for agents (bins + orphan config)
- [x] removal core (dry-run + backup) + `purge`
- [x] interactive list (select active items to remove)
- [x] `scan` for plugins/marketplaces
- [x] `scan` for MCP servers
- [x] full uninstall of an active package (`npm rm -g`, `uv tool uninstall`) and MCP (`claude mcp remove`)

## Development

- **How to work here (commands, safety, workflow):** [`docs/WORKING.md`](docs/WORKING.md)
- **Architecture and design:** [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)
- **Design decisions:** [`docs/adr/`](docs/adr/)

Local gate before committing: `./scripts/check.sh`
