# unipm — Universal Package Manager

> A meta package manager that unifies `apt`, `npm`, `pypi`, `flatpak`, `brew`,
> `appimage`, and `pacman`/`yay` (via Distrobox) under a single CLI.
> **⚠️ Current status: specification-only. Zero Go code exists. See
> [specs/index.md](specs/index.md) for the full spec suite.**

## Build & Development

> **Note**: No `go.mod` exists yet. The first implementation task is `go mod
> init github.com/DaviMGDev/unipm`.

```bash
# Initialize Go module (one-time)
go mod init github.com/DaviMGDev/unipm

# Install tooling
go install mvdan.cc/gofumpt@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Build
go build -o unipm ./cmd/unipm

# Run locally
./unipm --help

# Lint
golangci-lint run ./...

# Format
gofumpt -w .
```

## Testing

### Tier structure (from [ADR-0003](specs/adr/0003-testing-strategy.md))

| Tier | Scope | Command | Blocks merge? |
|------|-------|---------|---------------|
| **Tier 1 — Logic** | Adapter output parsing, flag construction, error handling, state/config, router dispatch, TUI model logic | `go test ./...` | Yes |
| **Tier 2 — Integration** | Real PMs in testcontainers (apt, npm, pip, brew, appimage, pacman/dnf) | `go test -tags=integration ./...` | Yes |
| **Tier 3 — E2E** | Full CLI workflows in VMs/Incus (flatpak, distrobox) | Nightly/manual only | No |

```bash
# Unit/logic tests (local, no Docker required)
go test ./...

# Integration tests (requires Docker)
go test -tags=integration ./...

# Single adapter integration test
go test -tags=integration -run TestApt ./pkg/adapter/

# With coverage (target: 80% per adapter package)
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test conventions

- **Co-located** test files: `pkg/adapter/apt_test.go`
- **Table-driven tests** with `testing.T` subtests
- **Golden-file fixtures** in `testdata/` for adapter output parsing
- **Isolation**: each test uses `t.TempDir()` for state/config — never depends on execution order
- **No real credentials**: mock `$HOME` for state/config paths
- **Flatpak & distrobox adapters** call `t.Skip()` when container capabilities are unavailable — the suite stays green

## Code Style

- **Language**: Go 1.22+
- **Formatting**: `gofumpt` (stricter than `gofmt`) — enforced in CI
- **Linting**: `golangci-lint` with project `.golangci.yml` (TODO: add config file)
- **Naming**:
  - Exported: `CamelCase`
  - Unexported: `camelCase`
  - Acronyms: all-caps (`HTTPClient`, `parseURL`)
  - File names: `snake_case.go`
- **Error handling**: never ignore errors. Wrap with context:
  ```go
  if err != nil {
      return fmt.Errorf("search %s: %w", pkg, err)
  }
  ```
- **Comments**: godoc comment on every exported function, type, and constant — starting with the identifier name
- **Logging**: `fmt.Fprintf(os.Stderr, ...)` for user-facing messages. No logging framework — unipm is a CLI, not a daemon.

## Architecture

### Pattern: Adapter (Plugin)

Every package manager backend implements the `PackageManager` interface. The
router holds a registry (`map[string]PackageManager`) and dispatches
polymorphically. See [ADR-0001](specs/adr/0001-adapter-pattern.md).

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

### Key design rules

- **No source ranking or categories** — all adapters are equal peers. The TUI lists sources alphabetically with no preselected default. ([ADR-0002](specs/adr/0002-no-source-priority.md))
- **Scoped sudo** — unipm itself never runs as root. Only the backend subprocess (e.g., `sudo apt install`) escalates.
- **No database** — all state is in `~/.unipm/state.json`; config is in `~/.unipm/config.yaml`.
- **Atomic state writes** — write to temp file, then rename. Never corrupt `state.json` on crash.

### Project structure (planned)

```
cmd/unipm/           # Main binary entrypoint (Cobra commands)
pkg/adapter/         # PackageManager interface + adapter implementations
pkg/router/          # Adapter registry and fan-out dispatch logic
pkg/state/           # state.json read/write with atomic writes
pkg/ui/              # Bubbletea TUI components (collision prompt only)
pkg/config/          # config.yaml parsing + default generation
specs/               # VibeSpecs specification files (see specs/index.md)
```

### Adding a new adapter

1. Create `pkg/adapter/<source>.go`
2. Implement the `PackageManager` interface
3. Register the adapter in the router's adapter list
4. Write tests in `pkg/adapter/<source>_test.go` using testcontainers
5. Update the backend matrix in `README.md`

## Spec-Driven Development

unipm uses **VibeSpecs** for specification-driven development. Every behavior
change starts in `specs/`, not in code.

### Reading order (see [specs/index.md](specs/index.md))

| # | File | Answers |
|---|------|---------|
| 1 | `specs/context.md` | Why? (problem, goals, stakeholders, scope) |
| 2 | `specs/users.md` | Who? (2 personas: Alex sysadmin, Jordan developer) |
| 3 | `specs/user_stories.md` | What? (6 stories US-001 to US-006 with EARS acceptance criteria) |
| 4 | `specs/features/*.feature` | How does each command behave? (Gherkin, 29 scenarios total) |
| 5 | `specs/architecture.md` | How is it structured? (data models, adapter contracts, state model) |
| 6 | `specs/adr/` | What did we decide and why? (3 ADRs: adapter pattern, no priority, testing strategy) |
| 7 | `specs/stack.md` | What tools? (Go, Cobra, Bubbletea, testcontainers) |
| 8 | `specs/ops.md` | How does it run? (distribution, config, roadmap 4 phases) |

### When to update specs

- New features → add/update `.feature` files and `user_stories.md`
- Architectural decisions → write an ADR in `specs/adr/` (use `0000-template.md`)
- Behavior changes → update existing `.feature` files
- Design rationale → update `architecture.md` or `context.md`

### Spec file rules

- **One concern per file.** Link out instead of duplicating.
- **ADR numbering**: sequential (`NNNN-brief-slug.md`).
- **Acceptance criteria**: EARS notation (5 patterns — see `user_stories.md`).
- **Decisions with trade-offs** belong in ADRs, not inline in code.

## Environment

- **Go**: 1.22+
- **Native package managers**: at least one of `apt`, `npm`, `pip3`, `flatpak`, `brew`, `pacman` for adapter testing
- **Docker**: required for Tier 2 integration tests (testcontainers-go)
- **Config files**: `~/.unipm/config.yaml` (auto-generated on first run), `~/.unipm/state.json` (tracks installs)
- **No external services**: all operations delegate to locally installed binaries. Network calls (npm registry, PyPI API, AppImageHub) are made directly by adapters — no third-party API dependencies.

## Dependencies (planned)

| Dependency | Version | Purpose |
|-----------|---------|---------|
| **Cobra** | 1.x | CLI framework (commands, flags, help, completion) |
| **Bubbletea** | 1.x | TUI for collision-resolution prompt only |
| **gofumpt** | latest | Stricter Go formatting |
| **golangci-lint** | latest | Aggregated linting |
| **testcontainers-go** | latest | Integration test isolation (Docker containers with real PMs) |

> **Note**: No `go.mod` exists yet. These are aspirational — versions will be
> pinned when `go mod init` runs.

## PR & Commit Guidelines

### ⚠️ Commit atomicity (MANDATORY — enforced by pre-commit hook)

**Commits MUST be small, atomic, and self-contained.** Each commit represents
exactly one logical change that can be reviewed, tested, and reverted in
isolation. This is the single most important rule in this project.

#### Hard limits (enforced by `.githooks/pre-commit`)

| Limit | Threshold | Action |
|-------|-----------|--------|
| Files changed | > 10 | ⚠️ Warning (reconsider) |
| Files changed | > 20 | 🚫 **REJECTED** |
| Lines added | > 500 | ⚠️ Warning (reconsider) |
| Lines added | > 1,000 | 🚫 **REJECTED** |

#### Rules

- **One logical change per commit.** If your commit message needs bullet points
to describe everything, it's too big. Split it.
- **Tests go in the same commit as the code they test** (never separate commits).
- **Format/lint fixes go with the change they belong to**, not as standalone commits
(unless they're project-wide and have zero behavior changes).
- **Renames/refactors MUST be their own commit** — never mix refactoring with
feature work. Renaming a type and adding a new feature in the same commit
makes `git blame` useless and bisect impossible.
- **Infrastructure-only commits are OK**: `.github/` workflows, `.githooks/`,
`.golangci.yml` can be standalone commits.
- **Never bundle unrelated packages.** `pkg/router/` + `pkg/ui/` + `pkg/state/`
renames in one commit is a violation.

#### Examples of ACCEPTABLE atomic commits

```
git add go.mod go.sum .gitignore && git commit -m "chore: init Go module and gitignore"
git add pkg/adapter/adapter.go && git commit -m "feat(adapter): define PackageManager interface"
git add pkg/config/config.go pkg/config/config_test.go && git commit -m "feat(config): add config load/save with defaults"
git add pkg/adapter/apt.go pkg/adapter/apt_test.go pkg/adapter/testdata/ && git commit -m "feat(adapter): add apt adapter with golden-fixture tests"
```

#### Examples of REJECTED commits

```
❌ feat(router,ui): extract router, add collision TUI, fix lint, rename types
   → 4 unrelated changes. Should be 4 commits.

❌ refactor: rename StateRecord→Record and add TUI and fix lint warnings
   → Rename + feature + lint all in one. Each goes in its own commit.

❌ Single commit touching 16 files across 5 packages
   → Rejected by pre-commit hook (files > 10, lines > 500).
```

This applies to AI agents and humans equally. Violating this rule produces
unreviewable history and makes `git bisect` useless.

### Commit format (Conventional Commits 1.0.0)

```
<type>(<scope>): <description>
```

**Types**: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, `style`, `perf`

**Scopes** (encouraged): `adapter`, `router`, `state`, `ui`, `config`, `cmd`, `specs`

**Breaking changes**: add `!` after type/scope + `BREAKING CHANGE:` footer.

**Examples**:
```
feat(adapter): add flatpak adapter with search and install
fix(router): deduplicate search results by (source, name)
docs: add CONTRIBUTING.md
feat(state)!: migrate state.json to v2 schema
```

### Branch naming

Use prefixes: `feat/`, `fix/`, `docs/`, `refactor/`, `test/`, `chore/`

### PR process

1. One logical change per PR
2. Update specs if behavior changes
3. Include tests (new features must have tests; bug fixes need regression tests)
4. All CI checks must pass (Tier 1 + Tier 2)
5. Squash-merge into `main` with a clean conventional-commit message

## Security

### Key rules

- **Never commit secrets** — no API keys, tokens, or credentials in the codebase
- **Scoped privilege escalation** — only backend subprocesses run with `sudo`; unipm binary never setuid
- **Atomic state writes** — write to temp file + rename prevents corruption
- **File permissions** — `~/.unipm/` created with `0700`, files with `0600`
- **Error messages sanitized** — no stack traces in user-facing output
- **Input validation** — validate `--source` flag against available adapters before dispatch

### Threat model summary

- **No network listeners, no daemon mode, no IPC, no plugin system** — attack surface is minimal
- **Supply chain risk** mitigated by delegating to native package managers (which have their own GPG/checksum verification)
- **Dependency scanning** via `govulncheck` and gitleaks (planned in CI)

## Implementation Roadmap

From [specs/ops.md](specs/ops.md):

| Phase | Scope | Target |
|-------|-------|--------|
| **Phase 1 — Skeleton** | Go module init, Cobra CLI scaffold, `PackageManager` interface, apt + npm adapters, `search` and `install` commands | `v0.1.0` |
| **Phase 2 — Router & State** | `state.json` read/write, `uninstall` and `update` commands, collision TUI (Bubbletea), `sources` command, `--source` flag | `v0.2.0` |
| **Phase 3 — Adapter Expansion** | pypi, flatpak, brew, appimage adapters; scoped sudo | `v0.3.0` |
| **Phase 4 — Distrobox & Polish** | Distrobox adapter, tab-completion, docs, CI, test suite | `v0.4.0` |
| **Stabilization** | Bug fixes, performance, beta releases | `v1.0.0-beta.x` |
| **GA** | First stable release | `v1.0.0` |

## Notes for AI Agents

- **Read specs before writing code.** All behavior is defined in `specs/`. Read `specs/index.md` for the reading order, then follow through all 8 spec files.
- **The `PackageManager` interface is the contract.** Every adapter must implement it — see [ADR-0001](specs/adr/0001-adapter-pattern.md). The router never changes when backends are added.
- **⚠️ Commits MUST be atomic (see hard limits above).** This is the #1 rule.
  - >10 files or >500 lines: split without asking.
  - Never bundle refactors + features + lint fixes in one commit.
  - Commit after EVERY logical step. No exceptions.
  - If you just completed 3 independent pieces of work, make 3 commits — not 1.
  - The pre-commit hook will REJECT commits >20 files or >1000 lines.
  - Tests go in the same commit as their code.
- **Tests run without Docker locally.** Tier 1 logic tests use golden-file fixtures — no containers required. Tier 2 integration tests require Docker and run in CI. See [ADR-0003](specs/adr/0003-testing-strategy.md).
- **State and config live in `~/.unipm/`.** Never hardcode paths. Use `os.UserHomeDir()` and create the directory on first access.
- **Specs use EARS notation** for acceptance criteria (5 patterns). The patterns are inlined in `specs/user_stories.md`.
- **Gherkin `.feature` files** in `specs/features/` define exact CLI behavior for every command. Use them as source of truth for implementation.
- **Breaking changes** require updating `state.json` version, writing a migration path, and a MAJOR semver bump. See `CHANGELOG.md` for versioning policy.
