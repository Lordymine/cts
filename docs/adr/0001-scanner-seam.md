# ADR 0001 — `Scanner` as the seam between the CLI and the categories

- **Status:** accepted
- **Date:** initial increment

## Context

cts needs to scan 4 distinct categories of junk (skills, agents, plugins, MCP), and each one lives in a different place and has its own "is it dead?" rules. The CLI needs to run them all and combine the results into a single report.

## Decision

Define a small `Scanner` interface (`Category()` + `Scan(ctx) ([]target.Target, error)`) in the `scan` package, and implement **one adapter per category** in `internal/scan/<category>`. `scan.Run` receives the scanners as a parameter and accumulates results and errors (`errors.Join`).

## Why (the 3 ADR criteria)

- **Hard to reverse:** the seam defines how every new category is plugged in; changing it later touches every adapter.
- **Surprising without context:** someone might expect a big `switch` over categories in main; choosing interface+adapters is deliberate.
- **Real trade-off:** interface+adapters (more files, extensible, testable in isolation) vs. a single file with everything (less ceremony, but couples the 4 logics and makes testing harder).

## Consequences

- Adding a category = creating an adapter and registering it in `main`. It doesn't touch the others.
- Each adapter is testable in isolation with `t.TempDir()`.
- The seam is justified by **real variation** (4 adapters), not hypothetical — it wouldn't be created for just 1.
