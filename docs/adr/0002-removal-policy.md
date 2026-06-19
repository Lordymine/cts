# ADR 0002 — Removal policy: dead removed automatically, active only explicitly, keepers protected

- **Status:** accepted
- **Date:** removal increment

## Context

cts deals with two distinct things:
1. **Dead / orphan / broken** — config with no binary, broken symlink. Proven junk; cts is certain it is safe to remove.
2. **Active / installed but unwanted** — e.g. `qwen` installed and working that the user no longer wants. cts **has no way to know on its own** that an active item is unwanted — that is the user's judgment.

Treating both the same would be dangerous (deleting something that works) or useless (cleaning only obvious junk, without replicating the real cleanup that removed active agents).

## Decision

- **`scan`** shows everything: dead items (marked `✗`) + active items (inventory).
- **`purge`** (automatic batch) removes **only the dead items** (`Dead == true`). It never touches an active item on its own.
- **Active items only by explicit selection.** An active (non-broken) item never enters `purge`. To remove one, the user **selects it from the list** (`cut <name>` or interactive selection) — a deliberate, item-by-item choice, with confirmation.
- **No hardcoded "protected" list.** cts has no way to know which agents are "keepers" — that is the user's knowledge, fragile and arbitrary. Protection of an active item comes from being **selection-only**; **backup** is the universal undo against accidents.
- **Dry-run is the default** in any removal; **backup** in `.cts-backups/` before deleting.

## Why (the 3 criteria)

- **Hard to reverse:** it defines the safety contract for all future removal.
- **Surprising without context:** someone might expect `purge` to delete everything `scan` listed; here `purge` is only the dead items, on purpose.
- **Real trade-off:** power (removing active items) vs. safety (not deleting what works without the user asking). Resolved by separating the **initiative**: cts acts on its own only on junk; the rest is an explicit choice.

## Consequences

- The `internal/remove` layer has **no protected list**: `purge` filters `Dead == true`; an active item leaves only by explicit name/selection.
- **Backup and dry-run are the removal layer's responsibility**, not the scanners' — the scanners stay read-only. Backup is what protects against accidental removal of an active item.
