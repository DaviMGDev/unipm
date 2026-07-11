# Changelog

All notable changes to unipm are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Types of changes: `Added`, `Changed`, `Deprecated`, `Removed`, `Fixed`, `Security`.

---

## [Unreleased]

### Added
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
