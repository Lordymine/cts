# ADR 0003 — Removal mechanisms: file + command

- **Status:** accepted
- **Date:** full-removal increment

## Context

Different categories "remove" in different ways:
- **dead skill / orphan config / plugin cache** → delete the file/dir.
- **active agent** → uninstall the package (`npm rm -g`, `bun rm -g`, `uv tool uninstall`) **and** delete the config.
- **MCP** → remove the config entry. The supported CLI `claude mcp remove <name> -s <scope>` does this correctly (it handles scope). Editing `~/.claude.json` by hand would reformat the entire file (195KB) and risk order/structure.

## Decision

Two mechanisms, both described in `Target`:
- **`Paths []string`** → files/dirs to delete.
- **`Uninstall []string`** → command to run **before** deleting (e.g. `["npm","rm","-g",pkg]`, `["claude","mcp","remove",name,"-s","user"]`).

The `Remover`: backs up the `Paths` → runs `Uninstall` (if any, via the injected `Runner`) → deletes the `Paths`. A target with no `Paths` **and** no `Uninstall` is skipped.

We do **not** edit JSON directly (`ConfigEdit` discarded) — `claude mcp remove` is the supported, safe path.

## Scope / limitations

- **MCP:** only the *user* scope gets an automatic command. A project server stays as inventory (removed inside the project, via cwd).
- **Agent with no known manager** (go-bin, python venv): removed by file only (config). The binary/install outside the `Dirs` is not touched in this version.

## Why (the 3 criteria)

- **Hard to reverse:** it defines how every removal happens.
- **Surprising without context:** one would expect a file delete for everything; command-for-some is deliberate.
- **Real trade-off:** supported command (`claude mcp remove`, `npm rm`) vs. editing config by hand. We chose the supported path — safer, less fragile.

## Consequences

- `Runner` injected into the `Remover` → testable (fake runner, executing nothing).
- Backup still happens before everything. In a command uninstall, the backup covers the `Paths` (config); the package reinstalls itself through its own manager.
