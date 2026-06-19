# Versioning

`cts` follows [Semantic Versioning 2.0.0](https://semver.org/): `MAJOR.MINOR.PATCH`.

- **MAJOR** — incompatible changes to commands, flags or behavior users rely on.
- **MINOR** — new functionality in a backward-compatible way (e.g. a new scanner).
- **PATCH** — backward-compatible bug fixes.

## Pre-1.0 (current phase)

While on `0.x`, the API and CLI surface are still stabilizing. Per SemVer, a **minor** bump (`0.MINOR.x`) may introduce breaking changes; **patch** bumps stay backward-compatible. Expect rapid iteration until `1.0.0`.

**Current version: `0.0.1`** — the initial public release.

## How a release happens

1. Update [`CHANGELOG.md`](CHANGELOG.md) with the new version and its changes.
2. Commit, then create an annotated tag:
   ```bash
   git tag -a v0.0.1 -m "v0.0.1"
   git push origin v0.0.1
   ```
3. Pushing a `v*` tag triggers the **Release** workflow (`.github/workflows/release.yml`), which cross-compiles binaries for Linux, macOS and Windows (amd64 + arm64) and publishes them on a GitHub Release.

The version is embedded into the binary at build time (`cts version`).

## Changelog

All notable changes are recorded in [`CHANGELOG.md`](CHANGELOG.md), following the [Keep a Changelog](https://keepachangelog.com/) format.
