# Prior Art — references and what we learned

Research before building: don't reinvent, don't get it wrong. Update this when you find a new relevant reference.

## Conclusion

No tool does the **unified** cleanup (skills + agents + plugins + MCP, multi-agent, with dry-run + backup) that `cts` does. The concept is ours. But there are mature pieces that **validate the architecture** and provide location/UX knowledge that would be easy to get wrong from scratch.

## References

### npkill — `voidcosmos/npkill` (~9.3k★, TypeScript)
Scans `node_modules` + lists size + interactive select + delete.
- **Architecture:** `/src` + `/tests`, almost zero dependencies, fast low-level scan.
- **Safety:** `--dry-run` flag, multi-select with preview, ⚠️ warning on a critical target. **Deletes permanently, no backup.**
- **What we take:** the scan→select→delete shape, the `--dry-run` flag, ⚠️ on a risky target, freed-space report.
- **What we do better:** dry-run is the **default** (not a flag), and **backup** before deleting.

### mcp-server-manager — `vlazic/mcp-server-manager` (Go)
Cross-platform single-binary MCP manager.
- **Architecture:** `cmd/` + `internal/` (+ `web/`) — validates our Go layout.
- **Config location (gold):** per client — `~/.claude.json`, `~/.gemini/settings.json`; MCP under the `mcpServers` key; different formats (`type:"http"` in Claude vs `httpUrl` in Gemini).
- **What we take:** the map of MCP config paths/formats per client, for our MCP scanner. Don't guess.

### Others (context)
- `radu2lupu/mcp-cleanup`, `YuancFeng/claude-code-cleanup`: clean up orphan **processes** — a different category (not our disk/config focus).
- `Guanff/claude-code-cleanup`: a scan→confirm→remove skill for session artifacts — confirms the pattern.

## Gaps (no direct prior art)
- **Detection of installed agents** (npm/bun/uv/go bins + orphan config): nobody covers it. Our own design — the list of installed items is **injected** (testable), with an orphan-config flag.
