# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.0.1] - 2026-06-18

Initial public release.

### Added
- **Scanners** for four categories of dead AI tooling:
  - skills (broken symlinks, missing `SKILL.md`),
  - agents (orphan config whose binary isn't installed),
  - plugins (orphan marketplaces with no installed plugin),
  - MCP servers (inventory + stdio servers with a missing command).
- **Removal core** with dry-run by default, backup to `.cts-backups/` before deleting, and an injected command runner. If the backup fails, nothing is deleted.
- **Proper uninstall**: packages via `npm rm -g` / `bun rm -g` / `uv tool uninstall`, MCP servers via `claude mcp remove`.
- **Commands**: `cts` (interactive menu), `cts scan`, `cts clean`, `cts purge [--yes]`, `cts help`, `cts version`.
- **Interactive UI** built with Charm `huh` + `lipgloss`: ASCII logo, menu, multi-select (dead pre-checked), and a color-coded report grouped by category.
- **Cross-platform config roots** (`internal/configroots`): agent configs are scanned across the home directory, `~/.config` (XDG), Windows `AppData`, and macOS `~/Library`.
- **CI** (vet, lint, race tests, build) and a tag-triggered **release** workflow building binaries for Linux, macOS and Windows.

[Unreleased]: https://github.com/Lordymine/cts/compare/v0.0.1...HEAD
[0.0.1]: https://github.com/Lordymine/cts/releases/tag/v0.0.1
