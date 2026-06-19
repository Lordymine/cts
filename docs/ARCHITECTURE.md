# Architecture — cts

Engineering document. Read it when touching design, adding a scanner, or trying to understand the flow.

## Overview

`cts` is a CLI that **scans** categories of junk on the machine and **removes** them safely. The flow is simple and linear:

```
main → scan.Run(scanners...) → []target.Target → report
                                              → cut/purge → backup → uninstall/delete
```

## Principle: deep modules, domain at the center

- **`internal/target`** — pure domain. Defines `Target` (what was found) and `Category`. It imports no IO. It is the stable center; everything points here.
- **`internal/scan`** — coordinates. Defines the `Scanner` interface (the **seam**) and `Run`, which runs every scanner and accumulates results/errors.
- **`internal/scan/<category>`** — one *adapter* per category (`skills`, then `agents`, `plugins`, `mcp`). Each one knows how to sweep its corner and decide what is dead. It hides that logic behind `Scan(ctx)`.
- **`main`** — wires the scanners with the real paths (`~/.claude/skills`, etc.) and prints. IO and wiring live at the edge.

Dependencies point inward: `skills → target`, `scan → target`, `main → scan, skills, target`. `target` imports nobody.

## Package layout

- `internal/target` — domain types (`Target`, `Category`).
- `internal/scan` — `Scanner` seam and `Run`.
- `internal/scan/<category>` — one adapter per category (`skills`, `agents`, `plugins`, `mcp`).
- `internal/configroots` — resolves OS-specific config base directories (see Cross-platform).
- `internal/dirsize` — directory size measurement.
- `internal/remove` — removal core (dry-run, backup, uninstall).
- `internal/ui` — presentation: logo, help, and formatted report. Render-only; no scan or removal logic.

## Cross-platform

The `internal/configroots` package resolves the OS-specific base directories where tool configs live. The `agents` scanner now looks for agent config dirs across all of them: the home directory (as `.<name>`), `~/.config/<name>` (Linux/macOS/XDG), Windows `AppData/Roaming` and `AppData/Local`, and macOS `~/Library/Application Support`. This makes agent-config coverage cross-platform; previously it only checked home dotdirs.

## The seam: `Scanner`

```go
type Scanner interface {
	Category() target.Category
	Scan(ctx context.Context) ([]target.Target, error)
}
```

A deliberately small interface (one work method). In Go, the adapter **does not declare** that it implements it — having the methods is enough. No inheritance, no `implements`: composition and implicit satisfaction.

Why an interface here and not direct code? Because there are **4 real adapters** coming (skills, agents, plugins, mcp). The seam is justified by real variation, not hypothetical. (If there were only 1, it would be premature abstraction — we wouldn't create it.)

## How to add a new scanner

1. Create `internal/scan/<category>/<category>.go` with a `Scanner` struct.
2. Implement `Category()` and `Scan(ctx) ([]target.Target, error)`.
3. Hide the "is it dead?" rule in a private function (`inspect`-style) — the caller doesn't need to know how it decides.
4. Write the **table-driven** test first, using `t.TempDir()` (never a real path).
5. Register it in `main` (`scan.Run(ctx, ..., newcategory.New(path))`).

## Errors

Errors are values. `Scan` returns an `error` wrapped with context (`fmt.Errorf("...: %w", err)`). `scan.Run` **accumulates** errors with `errors.Join` and keeps going — one broken scanner doesn't bring down the others. A nonexistent directory is **not an error** (there is simply nothing to clean).

## Safety

- Dry-run by default; removal only with an explicit flag + confirmation.
- Backup in `.cts-backups/` before deleting.
- Removal lives in `internal/remove` with the same discipline: small surface, testable with `t.TempDir()` and an injected `Runner`. See ADR 0003 for the file-vs-command mechanisms.

## Known limitations

- Agents without a known package manager (go-installed binaries, Python venvs) are removed by config file only; the binary outside the config dirs is not touched.
- MCP: only user-scope servers get an automatic removal command; project-scope servers are inventory-only.
