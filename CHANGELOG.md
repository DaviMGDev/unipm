# Changelog

All notable changes to unipm are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Types of changes: `Added`, `Changed`, `Deprecated`, `Removed`, `Fixed`, `Security`.

---

## [Unreleased]

### Added
- Distrobox adapter with multi-PM support (apt, pacman, yay, dnf, zypper).
- Cache-based tab-completion for package names and `--source` flag.
- Integration test stubs for pypi, flatpak, brew, and appimage adapters.

### Changed
- CI workflow: coverage threshold is now a hard fail (was warning).
- CI workflow: added skipped-adapter tracking in Tier 2 job.

### Fixed
- URL-encode npm search queries to prevent malformed HTTP requests.
- `go mod tidy` to properly mark direct dependencies (cobra, bubbletea, yaml).

### Security
- Added `.gitleaks.toml` configuration for secret detection.

---

## [0.3.0] — 2026-07-11

### Added
- **PyPI adapter** (`pypi`): search via PyPI JSON API, install via `pip3 install --user`,
  uninstall via `pip3 uninstall -y`, info via `pip3 show`. Golden-fixture tests.
- **Flatpak adapter** (`flatpak`): search/install/uninstall/info via `flatpak` CLI.
  Golden-fixture tests. Graceful skip in Docker environments.
- **Brew adapter** (`brew`): search/install/uninstall via `brew` CLI, info via
  `brew info --json`. Golden-fixture tests.
- **AppImage adapter** (`appimage`): search via AppImageHub API, install via
  download to `~/Applications`, uninstall via file removal. Golden-fixture tests.

---

## [0.2.0] — 2026-07-11

### Added
- **Router package** (`pkg/router`): adapter registry with parallel search dispatch,
  deduplication by (source, name), and per-adapter timeouts.
- **State package** (`pkg/state`): atomic read/write of `~/.unipm/state.json`
  with version validation. CRUD operations: Add, Remove, Get, List, UpdateVersion.
- **Uninstall command**: looks up package in state.json, routes to correct backend,
  offers clean-state option on backend failure.
- **Update command**: refreshes all tracked packages or a single named package.
  Partial failure handling with per-package reporting.
- **Collision TUI** (`pkg/ui`): interactive Bubbletea prompt when a package
  exists in multiple sources. Vim-key navigation (j/k/↑/↓), alphabetical listing
  per ADR-0002, no preselected default.
- **Sources command**: lists all compiled-in adapters with ✓/✗ availability.
- **`--source` flag**: explicit adapter selection with comma-separated multi-source
  support. Validation against available adapters.

### Changed
- Search and install commands now dispatch through the router instead of
  iterating adapters directly.

---

## [0.1.0] — 2026-07-10

### Added
- Project bootstrap: Go module, `.gitignore`, `.golangci.yml`, CI workflows
  (test, lint, security).
- **PackageManager interface** (`pkg/adapter`): defines the 6-method contract
  for all backends (Name, Search, Install, Uninstall, Info, IsAvailable).
- **Config package** (`pkg/config`): YAML config with distrobox definitions,
  cache TTL, and search timeout. Auto-generated defaults on first run.
- **State package** (`pkg/state`): JSON state file with atomic writes.
- **Apt adapter** (`apt`): search via `apt search`, install via `sudo apt install -y`,
  uninstall via `sudo apt remove -y`, info via `apt show`.
- **Npm adapter** (`npm`): search via npm registry API, install via `npm install -g`,
  uninstall via `npm uninstall -g`, info via `npm info --json`.
- **Cobra CLI scaffold**: root command, search, install, install --source,
  completion shell scripts.
- Golden-file test fixtures for apt and npm adapters.
- `AGENTS.md` with AI agent instructions and commit rules.
- `plan.md` and `tasks.md` with implementation roadmap.
- Project specification suite (VibeSpecs): context, personas, user stories,
  architecture, stack, operations, ADRs, and Gherkin feature files.
- `README.md` with command reference, backend matrix, and quickstart guide.
- `CONTRIBUTING.md` with development workflow and conventions.
- `CHANGELOG.md` (this file) following Keep a Changelog.
- `LICENSE` (MIT).

---

## Versioning

unipm follows **Semantic Versioning 2.0.0** (`MAJOR.MINOR.PATCH`).

### How Versions Are Bumped

| Change type | Segment | Example |
|-------------|---------|---------|
| Breaking CLI flag changes, config/state schema changes requiring user migration, removed adapters, Go API breakage in public `pkg/` packages | **MAJOR** | `1.0.0 → 2.0.0` |
| New adapter, new subcommand, new `--flag`, non-breaking config additions, new state.json fields (with backward-compatible defaults) | **MINOR** | `1.0.0 → 1.1.0` |
| Bug fixes, performance improvements, documentation updates | **PATCH** | `1.0.0 → 1.0.1` |

### Pre-release Versions

Before `1.0.0`, the `MINOR` segment tracks feature milestones and `PATCH`
tracks fixes. Pre-release tags may be used for beta/RC builds:

```
v0.1.0          # First working skeleton
v0.2.0          # Router + state + collision TUI
v0.3.0          # All adapters
v0.4.0          # Distrobox + completion
v1.0.0-beta.1   # Feature-complete, stabilization phase
v1.0.0          # First stable release
```

### State File Versioning

The `state.json` file carries its own `"version"` field (starting at `1`),
independent of the unipm release version. If the state schema changes
incompatibly, unipm increments the state version and rejects unknown versions
with a migration instruction — it never silently corrupts data. The state
version is NOT tied to semver.

### Config Compatibility

`config.yaml` keys are **additive**. New keys must include sensible defaults so
existing configs work without modification. Removing or renaming a key requires
a MAJOR bump.

### Release Cadence

No fixed schedule. Versions are released when milestones (see `specs/ops.md`
roadmap) are complete:

| Milestone | Expected version |
|-----------|-----------------|
| Phase 1 — Skeleton (CLI + 2 adapters) | `0.1.0` |
| Phase 2 — Router & State (uninstall, update, TUI) | `0.2.0` |
| Phase 3 — Adapter expansion (7 backends) | `0.3.0` |
| Phase 4 — Distrobox & polish | `0.4.0` |
| Stabilization + beta | `1.0.0-beta.x` |
| First stable release | `1.0.0` |

### Tag Format

Use annotated tags with the `v` prefix:

```bash
git tag -a v0.1.0 -m "v0.1.0: initial skeleton — cobra CLI, apt + npm adapters"
git push origin v0.1.0
```

GitHub Releases are created from these tags by CI (GoReleaser). Release notes
are generated from the corresponding `CHANGELOG.md` section.
