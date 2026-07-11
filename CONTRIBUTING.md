# Contributing to unipm

Thank you for your interest in contributing! unipm is a meta package manager
that unifies `apt`, `npm`, `pip`, `flatpak`, `brew`, `pacman`, and distrobox
under a single CLI. This document covers everything you need to start
contributing.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Coding Conventions](#coding-conventions)
- [Commit Conventions](#commit-conventions)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Specifications](#specifications)
- [Release Process](#release-process)

---

## Code of Conduct

Be respectful. Assume good intent. Give constructive feedback. This project
follows the [Contributor Covenant](https://www.contributor-covenant.org/).

---

## Getting Started

### Prerequisites

- **Go 1.22+** — the entire codebase is Go.
- **golangci-lint** — aggregated linting (run before pushing).
- **gofumpt** — stricter Go formatting (enforced in CI).
- **At least one native package manager** — `apt`, `npm`, `pip3`, `flatpak`,
  `brew`, or `pacman` — to test the adapter you're working on.
- **Docker** (for integration tests) — testcontainers-go spins up isolated
  containers with real package managers.

### Setup

```bash
# Clone
git clone https://github.com/DaviMGDev/unipm.git
cd unipm

# Install tooling
go install mvdan.cc/gofumpt@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Build
go build -o unipm ./cmd/unipm

# Verify
./unipm --version

# Enable pre-commit hooks (enforces atomic commits + conventional commit format)
git config core.hooksPath .githooks
```

---

## Development Workflow

1. **Fork** the repository.
2. **Create a feature branch** from `main`:
   ```bash
   git checkout -b feat/my-feature
   ```
   Use branch prefixes: `feat/`, `fix/`, `docs/`, `refactor/`, `test/`, `chore/`.
3. **Make your changes** following the [coding conventions](#coding-conventions).
4. **Run tests and linting** before committing:
   ```bash
   go test ./...
   golangci-lint run ./...
   ```
5. **Commit** using [conventional commits](#commit-conventions).
6. **Push** to your fork and open a pull request against `main`.

---

## Coding Conventions

### Go Style

- **Formatting**: `gofumpt` on save (stricter than `gofmt`).
- **Linting**: `golangci-lint` with the project's `.golangci.yml` (TODO: add
  config file; currently uses defaults).
- **Naming**:
  - Exported identifiers: `CamelCase`.
  - Unexported: `camelCase`.
  - Acronyms are all-caps: `HTTPClient`, `parseURL`.
  - File names: `snake_case.go`.
- **Error handling**: Never ignore errors. Wrap with context:
  ```go
  if err != nil {
      return fmt.Errorf("search %s: %w", pkg, err)
  }
  ```
- **Comments**: Document every exported function, type, and constant with a
  godoc comment starting with the identifier name.
- **Logging**: Use `fmt.Fprintf(os.Stderr, ...)` for user-facing messages.
  unipm does not depend on a logging framework — it's a CLI tool, not a daemon.

### Project Structure

```
cmd/unipm/           # Main binary entrypoint (Cobra commands)
pkg/adapter/         # PackageManager interface + adapter implementations
pkg/router/          # Adapter registry and fan-out dispatch logic
pkg/state/           # state.json read/write with atomic writes
pkg/ui/              # Bubbletea TUI components (collision prompt only)
pkg/config/          # config.yaml parsing + default generation
specs/               # VibeSpecs specification files (see specs/index.md)
```

### Adding a New Adapter

1. Create a file `pkg/adapter/<source>.go`.
2. Implement the `PackageManager` interface:
   ```go
   type PackageManager interface {
       Name() string
       Search(query string) ([]Package, error)
       Install(pkg Package) error
       Uninstall(pkg Package) error
       Info(pkg Package) (Details, error)
       IsAvailable() bool
   }
   ```
3. Register the adapter in the router's adapter list.
4. Write tests in `pkg/adapter/<source>_test.go` using testcontainers.
5. Add the adapter to the backend matrix in `README.md`.

See [ADR-0001](specs/adr/0001-adapter-pattern.md) for the rationale behind the
adapter pattern.

---

---

## Commit Atomicity

> ⚠️ **This is the most important rule in the project.** Enforced by
> `.githooks/pre-commit`.

**Commits MUST be small, atomic, and self-contained.** Each commit represents
exactly one logical change that can be reviewed, tested, and reverted in
isolation.

### Hard limits

| Limit | Threshold | Action |
|-------|-----------|--------|
| Files changed | > 10 | ⚠️ Warning |
| Files changed | > 20 | 🚫 Rejected (blocks commit) |
| Lines added | > 500 | ⚠️ Warning |
| Lines added | > 1,000 | 🚫 Rejected (blocks commit) |

### Rules

- **One logical change per commit.** If your commit message needs bullet points,
split it.
- **Tests go in the same commit as the code they test.**
- **Renames/refactors MUST be their own commit.** Never mix refactoring with features.
- **Never bundle unrelated packages.** `pkg/router/` + `pkg/ui/` + lint fixes
in one commit is a violation.

### Enabling the hooks

```bash
git config core.hooksPath .githooks
```

The hooks are automatically active after `git clone` if you run the setup
above. CI also validates commit conventions.

---

## Commit Conventions

This project follows **Conventional Commits** (`1.0.0`). Every commit message
must use one of these types:

| Type | When to use |
|------|-------------|
| `feat` | A new feature (user-facing or adapter) |
| `fix` | A bug fix |
| `docs` | Documentation only (README, specs, comments) |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `test` | Adding or updating tests |
| `chore` | Build tasks, CI, dependencies, tooling |
| `style` | Formatting, whitespace (no code change) |
| `perf` | Performance improvement |

### Format

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

**Scope** is optional but encouraged — use the package or component name (e.g.,
`adapter`, `router`, `state`, `ui`, `config`, `cmd`, `specs`).

**Examples**:

```
feat(adapter): add flatpak adapter with search and install
```

```
fix(router): deduplicate search results by (source, name)
```

```
docs: add CONTRIBUTING.md
```

Breaking changes: add `!` after the type/scope and a `BREAKING CHANGE:` footer:

```
feat(state)!: migrate state.json to v2 schema

BREAKING CHANGE: state.json v1 files are no longer supported.
Users must run `unipm migrate` to upgrade.
```

---

## Testing

### Tiers

| Tier | Scope | Command | Blocks merge? |
|------|-------|---------|---------------|
| **Unit** | Individual adapters, state, config, router | `go test ./...` | Yes |
| **Integration** | Adapter ↔ real PM via testcontainers | `go test -tags=integration ./...` | Yes |
| **E2E** | Full CLI workflows | Shell scripts + containers | No (alert only) |

### Conventions

- Test files are co-located: `pkg/adapter/apt_test.go`.
- Use **table-driven tests** (`testing.T` with subtests).
- Each adapter must have at least one happy-path install/uninstall integration
  test.
- Tests must not depend on execution order. Each test creates its own temp
  directory for state and config files.
- Use golden-file patterns (`testdata/`) for adapter output parsing fixtures.
- Never hardcode real credentials or home-directory paths. Use `t.TempDir()`.

### Running Tests

```bash
# Unit tests
go test ./...

# Integration tests (requires Docker)
go test -tags=integration ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

Coverage target: **80% per adapter package** (enforced in CI).

---

## Pull Request Process

1. **Check existing issues and PRs** to avoid duplication.
2. **Keep PRs focused** — one logical change per PR.
3. **Update specs if applicable** — if your change affects behavior, update or
   add files in `specs/` (`.md` for design, `.feature` for behavior).
4. **Include tests** — new features must include tests; bug fixes should include
   a regression test.
5. **Pass CI** — all status checks must be green before review.
6. **Request review** from a maintainer.
7. **Address feedback** — be open to suggestions; the maintainer may ask for
   changes.
8. **Squash merge** — the maintainer will squash-merge into `main` with a
   clean conventional-commit message.

### PR Title Convention

Use the same conventional-commit format as commits. The PR title becomes the
squash-merge commit message:

```
feat(adapter): add brew adapter with search, install, and uninstall
```

---

## Specifications

unipm uses **VibeSpecs** for specification-driven development. All specs live
in the `specs/` directory. See [`specs/index.md`](specs/index.md) for the full
reading order and file conventions.

### When to Update Specs

- New features → add or update `.feature` files and `user_stories.md`.
- Architectural decisions → write an ADR in `specs/adr/`.
- Behavior changes → update existing `.feature` files.
- Design rationale → update `architecture.md` or `context.md`.

### Spec File Rules

- **One concern per file.** Link out instead of duplicating.
- **ADR numbering**: sequential (`NNNN-brief-slug.md`). Use `0000-template.md`
  as the template.
- **Acceptance criteria**: use EARS notation (5 patterns).
- **Decisions with trade-offs** belong in ADRs, not inline in code.

---

## Release Process

See [`CHANGELOG.md`](CHANGELOG.md) for the version history. unipm follows
[Semantic Versioning](https://semver.org) (`MAJOR.MINOR.PATCH`).

### Releasing a New Version

1. Ensure `main` is green (all CI checks pass).
2. Update `CHANGELOG.md` — move `[Unreleased]` entries to a new version section.
3. Bump the version in the code (e.g., a `version` constant in `cmd/unipm/` or
   via `-ldflags` at build time).
4. Create an annotated tag:
   ```bash
   git tag -a v1.0.0 -m "v1.0.0: first stable release"
   ```
5. Push the tag:
   ```bash
   git push origin v1.0.0
   ```
6. CI builds platform binaries via GoReleaser and attaches them to the GitHub
   Release.
7. Publish the GitHub Release with the changelog entries as the release notes.

### Version Compatibility

- `state.json` includes a `"version"` field. If unipm encounters an unknown
  version, it errors with an upgrade message — never silently corrupts data.
- The `config.yaml` schema is additive — new keys must have sensible defaults so
  older configs continue to work.

---

## Questions?

Open an issue on GitHub or start a discussion. We're happy to help newcomers get
started.

---

## See Also

- [`README.md`](README.md) — project overview and command reference
- [`specs/index.md`](specs/index.md) — spec reading order
- [`CHANGELOG.md`](CHANGELOG.md) — version history
- [ADR-0001](specs/adr/0001-adapter-pattern.md) — why the adapter pattern
- [ADR-0002](specs/adr/0002-no-source-priority.md) — why no source ranking
