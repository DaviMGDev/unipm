# Implementation Plan — unipm

## Goal

Deliver a working meta package manager CLI that unifies `apt`, `npm`, `pypi`,
`flatpak`, `brew`, `appimage`, and `pacman`/`yay` (via Distrobox) under a single
`unipm` command. Every user story (US-001 through US-006) must be satisfied and
every Gherkin scenario in `specs/features/` must pass. The implementation
proceeds in 4 sequential phases, each producing a tagged release.

Success criteria:
- `unipm search <query>` returns deduplicated results from all available adapters in under 10 s
- `unipm install <pkg>` installs directly when unambiguous, opens a TUI when ambiguous
- `unipm uninstall <pkg>` routes to the correct backend using `state.json`
- `unipm update` refreshes all tracked packages
- `unipm sources` lists adapter availability
- `unipm completion <shell>` generates working completion scripts
- `go test ./...` passes (Tier 1, zero containers) with ≥80% adapter package coverage
- `go test -tags=integration ./...` passes (Tier 2, Docker) in CI

---

## Requirements

### Functional

| ID | Capability | User story | Feature file |
|----|-----------|------------|--------------|
| F1 | Search all backends in parallel | US-001 | `search.feature` (5 scenarios) |
| F2 | Install with collision TUI | US-002 | `install.feature` (7 scenarios) |
| F3 | Uninstall via state lookup | US-003 | `uninstall.feature` (4 scenarios) |
| F4 | Update tracked packages | US-004 | `update.feature` (4 scenarios) |
| F5 | List available sources | US-005 | `sources.feature` (3 scenarios) |
| F6 | Shell tab-completion | US-006 | `completion.feature` (6 scenarios) |
| F7 | 7 backend adapters (apt, npm, pypi, flatpak, brew, appimage, distrobox) | US-001, US-002 | `search.feature`, `install.feature` |
| F8 | Atomic state writes to `state.json` | US-002, US-003, US-004 | `install.feature`, `uninstall.feature`, `update.feature` |
| F9 | Config parsing (`config.yaml` with distrobox, cache_ttl, search_timeout) | US-001, US-006 | `sources.feature`, `completion.feature` |
| F10 | `--source` flag for explicit adapter selection | US-002 | `install.feature` |

### Non-functional

- **Performance**: `unipm sources` < 100 ms; `search` results start < 2 s; tab-completion < 50 ms cached
- **Reliability**: atomic state writes (no corruption); graceful adapter timeout (partial results)
- **Security**: scoped sudo (only backend subprocesses); `~/.unipm/` with 0700/0600; no secrets in code; `govulncheck` + `gitleaks` in CI
- **Coverage**: ≥80% line coverage per adapter package (Tier 1)
- **Binary**: single statically-compiled Go binary, no external runtime dependency

### Acceptance Criteria

1. `unipm search htop` run on a system with apt + brew → table output with Source, Name, Version, Description columns and one row per backend
2. `unipm install requests` on a system where both apt and pypi have "requests" → TUI opens, user selects a source, package installs, record appears in `state.json`
3. `unipm install htop --source apt` on a system with apt → installs directly, no prompt, record in `state.json`
4. `unipm uninstall htop` where htop was installed via unipm → delegates to correct backend, removes record from `state.json`
5. `unipm uninstall nonexistent` where package not in `state.json` → error "not installed via unipm", non-zero exit
6. `unipm sources` → lists every compiled-in adapter with ✓ or ✗ status
7. `unipm completion bash | source` then `<TAB>` after `unipm install --source ` → completes with available adapter names

---

## Context

### Files to create

```
.gitignore
.golangci.yml
.github/workflows/test.yml
.github/workflows/lint.yml
.github/workflows/security.yml
go.mod
go.sum
cmd/unipm/main.go
cmd/unipm/root.go
cmd/unipm/search.go
cmd/unipm/install.go
cmd/unipm/uninstall.go
cmd/unipm/update.go
cmd/unipm/sources.go
cmd/unipm/completion.go
pkg/adapter/adapter.go             # PackageManager interface + Package struct
pkg/adapter/apt.go
pkg/adapter/npm.go
pkg/adapter/pypi.go
pkg/adapter/flatpak.go
pkg/adapter/brew.go
pkg/adapter/appimage.go
pkg/adapter/distrobox.go
pkg/adapter/apt_test.go
pkg/adapter/npm_test.go
pkg/adapter/pypi_test.go
pkg/adapter/flatpak_test.go
pkg/adapter/brew_test.go
pkg/adapter/appimage_test.go
pkg/adapter/distrobox_test.go
pkg/adapter/testdata/apt_search_htop.txt
pkg/adapter/testdata/apt_search_multi.txt
pkg/adapter/testdata/npm_search_htop.json
pkg/adapter/testdata/pypi_search_htop.json
pkg/adapter/testdata/flatpak_search_htop.txt
pkg/adapter/testdata/brew_search_htop.txt
pkg/adapter/testdata/appimage_search_htop.json
pkg/router/router.go
pkg/router/router_test.go
pkg/state/state.go
pkg/state/state_test.go
pkg/ui/tui.go
pkg/ui/tui_test.go
pkg/config/config.go
pkg/config/config_test.go
```

### Dependencies

| Module | Version | Purpose |
|--------|---------|---------|
| `github.com/spf13/cobra` | v1.8.1 | CLI framework |
| `github.com/charmbracelet/bubbletea` | v1.1.0 | TUI collision prompt |
| `github.com/charmbracelet/lipgloss` | v0.13.0 | Terminal styling (tables) |
| `github.com/testcontainers/testcontainers-go` | v0.31.0 | Integration test containers (Tier 2) |
| `mvdan.cc/gofumpt` | v0.6.0 | Formatter (dev tool) |
| `github.com/golangci/golangci-lint` | v1.59.0 | Linter (dev tool) |

All versions will be pinned when `go mod init` runs and `go mod tidy` resolves the graph.

### References

- `specs/context.md` — problem, scope, goals, non-goals
- `specs/users.md` — 2 personas (Alex, Jordan), 2 journeys
- `specs/user_stories.md` — 6 stories with EARS acceptance criteria
- `specs/features/*.feature` — 29 Gherkin scenarios
- `specs/architecture.md` — data models, adapter contracts, state model
- `specs/adr/0001-adapter-pattern.md` — why the adapter pattern
- `specs/adr/0002-no-source-priority.md` — why no source ranking
- `specs/adr/0003-testing-strategy.md` — tiered testing strategy
- `specs/stack.md` — technology choices
- `specs/ops.md` — roadmap, config, security

### Constraints

- Go 1.22+ (`gofumpt` formatting, `golangci-lint` linting)
- Conventional commits (`feat(scope): description`)
- Squash-merge into `main`
- No spec changes — specs are source of truth; implementation conforms to them
- Tier 1 tests must pass without Docker (golden-file fixtures)
- Tier 2 tests require Docker, run in CI, may `t.Skip()` for flatpak/distrobox

---

## Out of Scope

- **Project-local dependency management** (`--dev`, `--save`) — deferred per `specs/context.md` non-goal #1
- **Graphical UI** — TUI handles the only interactive need; no GUI planned per non-goal #2
- **Daemon/service mode** — local CLI only per non-goal #3
- **macOS/Windows support** — Linux-only in v1 per non-goal #4
- **User-configurable source priority** — rejected in ADR-0002; may revisit post-v1.0
- **Package manager protocol reimplementation** — unipm delegates, does not replace native PMs
- **Snap backend** — not in the 7 specified adapters

---

## Design / Architecture

### Key decisions

1. **Adapter pattern** (ADR-0001): Every backend implements `PackageManager` interface. Router holds `map[string]PackageManager`. Adding a backend = one file + registration. Router never changes.

2. **No source ranking** (ADR-0002): Collision TUI lists sources alphabetically, no preselected default. `--source` flag bypasses TUI for users who know their preference.

3. **Tiered testing** (ADR-0003): Tier 1 (logic, golden fixtures, local), Tier 2 (integration, Docker, CI), Tier 3 (E2E, VMs, nightly). Only Tier 1 runs locally. Flatpak/distrobox skip gracefully in Tier 2 if environment lacks capabilities.

4. **Scoped sudo**: unipm binary never runs as root. Only backend subprocesses (e.g., `sudo apt install`) escalate.

5. **No database**: State in `~/.unipm/state.json`, config in `~/.unipm/config.yaml`, cache in `~/.unipm/cache.json`. Atomic writes via write-to-temp-then-rename.

### Package dependency order

```
pkg/config     (zero internal deps — reads config.yaml)
pkg/state      (zero internal deps — reads/writes state.json)
pkg/adapter    (zero internal deps — defines interface + structs)
pkg/router     (depends on pkg/adapter)
pkg/ui         (depends on pkg/adapter)
cmd/unipm/*    (depends on pkg/router, pkg/state, pkg/config, pkg/ui, pkg/adapter)
```

Adapters (`pkg/adapter/apt.go`, etc.) depend only on the `PackageManager` interface — never on each other, never on the router.

### Phase sequencing

Phases are sequential — each depends on the previous:

```
Phase 1 (Skeleton)
  └─→ Phase 2 (Router & State)
        └─→ Phase 3 (Adapter Expansion)
              └─→ Phase 4 (Distrobox & Polish)
```

Within Phase 1: `go mod init` → `.gitignore` / `.golangci.yml` / CI → interface → config → state → first two adapters → Cobra scaffold → search → install.

Within Phase 2: router → state.json integration → uninstall → update → collision TUI → sources command → `--source` flag.

Within Phase 3: pypi → flatpak → brew → appimage adapters (each with Tier 1 golden fixtures + Tier 2 integration tests).

Within Phase 4: distrobox adapter → tab-completion → docs polish → CI hardening.

### Open questions

- None. All architectural questions are resolved by the 3 ADRs. Implementation details (e.g., exact Cobra command structure, Bubbletea model shape) are standard Go patterns and will be decided during implementation.

---

## Implementation Steps

### Phase 1 — Skeleton (target: v0.1.0)

**Step 1: Project bootstrap**
Files: `.gitignore`, `go.mod`, `.golangci.yml`, `.github/workflows/test.yml`, `.github/workflows/lint.yml`, `.github/workflows/security.yml`
Depends on: nothing
Creates the Go module, gitignore with `unipm` binary / `.DS_Store` / `vendor/` / `*.out` / `coverage.out`, golangci-lint config, and 3 CI workflows (test with Tier 1 + Tier 2, lint with gofumpt + golangci-lint, security with gitleaks + govulncheck).

**Step 2: PackageManager interface + Package struct**
Files: `pkg/adapter/adapter.go`
Depends on: Step 1
Defines `Package` struct, `Details` struct, and `PackageManager` interface (6 methods) per `specs/architecture.md`.

**Step 3: Config package**
Files: `pkg/config/config.go`, `pkg/config/config_test.go`
Depends on: Step 1
Parses `~/.unipm/config.yaml` with `DistroboxConfig`, `CacheTTL` (default 86400), `SearchTimeout` (default 10). Auto-generates defaults on first run. Tests with fixture YAML in `t.TempDir()`.

**Step 4: State package**
Files: `pkg/state/state.go`, `pkg/state/state_test.go`
Depends on: Step 1
Reads/writes `~/.unipm/state.json` with atomic write (temp file + rename). Validates version field. Methods: `Add()`, `Remove()`, `Get()`, `List()`, `UpdateVersion()`. Tests with isolated `t.TempDir()`.

**Step 5: Apt adapter**
Files: `pkg/adapter/apt.go`, `pkg/adapter/apt_test.go`, `pkg/adapter/testdata/apt_search_htop.txt`
Depends on: Step 2
Implements `PackageManager` for apt. `Search()` runs `apt search`, parses output. `Install()` runs `sudo apt install -y`. `Uninstall()` runs `sudo apt remove -y`. `IsAvailable()` checks `$PATH` for `apt`. Tier 1 tests with golden-file fixture for search output parsing.

**Step 6: Npm adapter**
Files: `pkg/adapter/npm.go`, `pkg/adapter/npm_test.go`, `pkg/adapter/testdata/npm_search_htop.json`
Depends on: Step 2
Implements `PackageManager` for npm. `Search()` uses npm registry API (`https://registry.npmjs.org/-/v1/search?text=`). `Install()` runs `npm install -g`. `Uninstall()` runs `npm uninstall -g`. Tier 1 tests with golden-file JSON fixture.

**Step 7: Cobra CLI scaffold**
Files: `cmd/unipm/main.go`, `cmd/unipm/root.go`
Depends on: Steps 3, 4
Sets up Cobra root command with `--version` (via `-ldflags`), persistent flags, and help template. Subcommands registered as stubs: `search`, `install`, `uninstall`, `update`, `sources`, `completion`.

**Step 8: Search command**
Files: `cmd/unipm/search.go`
Depends on: Steps 5, 6, 7
Wires `search` command. Reads `search_timeout` from config. Iterates available adapters, calls `Search()` in parallel via goroutines, merges results, deduplicates by `(Source, Name)`, outputs table with `lipgloss`. Matches `specs/features/search.feature` scenarios.

**Step 9: Install command (single-match path)**
Files: `cmd/unipm/install.go`
Depends on: Steps 7, 8
Implements install with single-backend match path (no TUI yet — that's Phase 2). Searches all backends, if exactly one has package, installs directly and records in `state.json`. `--source` flag skips search and installs from named adapter. Matches `install.feature` single-match and `--source` scenarios.

**Phase 1 verification**: `go build -o unipm ./cmd/unipm` succeeds. `./unipm --help` shows command tree. `./unipm search htop` returns results if apt or npm has htop. `./unipm install htop --source apt` installs and records.

---

### Phase 2 — Router & State (target: v0.2.0)

**Step 10: Router package**
Files: `pkg/router/router.go`, `pkg/router/router_test.go`
Depends on: Steps 2, 5, 6
Registry: `map[string]PackageManager`. `Register()` adds adapter. `Get()` retrieves by name. `ListAvailable()` returns names of registered adapters. `SearchAll()` fans out to all adapters in parallel, merges, deduplicates. Tests with mock adapters.

**Step 11: Wire router into commands**
Files: `cmd/unipm/search.go` (refactor), `cmd/unipm/install.go` (refactor), `cmd/unipm/root.go`
Depends on: Step 10
Replace direct adapter calls in search/install with router dispatch. Router registry populated at startup from available adapters only.

**Step 12: Uninstall command**
Files: `cmd/unipm/uninstall.go`
Depends on: Steps 4, 10, 11
Looks up package in `state.json`, routes to recorded source's adapter, removes record on success. If not tracked: error "was not installed via unipm". If backend fails: show native error, offer to clean state record. Matches `specs/features/uninstall.feature`.

**Step 13: Update command**
Files: `cmd/unipm/update.go`
Depends on: Steps 4, 10, 11
Without args: iterates all records in `state.json`, delegates update to each adapter, refreshes version. With arg: updates single package. Partial failure: reports per-package, continues, non-zero exit. Matches `specs/features/update.feature`.

**Step 14: Collision TUI**
Files: `pkg/ui/tui.go`, `pkg/ui/tui_test.go`
Depends on: Step 2
Bubbletea model: lists matching sources alphabetically (no preselected default), j/k/↑/↓ to navigate, Enter to select, q/Esc to cancel. Returns selected `Package`. Tests verify model state transitions (no terminal required).

**Step 15: Wire TUI into install**
Files: `cmd/unipm/install.go` (refactor)
Depends on: Steps 14, 11
When multiple backends have the package and no `--source` flag given, open TUI. When user selects, install and record. Matches `install.feature` collision scenario.

**Step 16: Sources command**
Files: `cmd/unipm/sources.go`
Depends on: Step 10
Lists all compiled-in adapters with ✓/✗ availability. Distrobox adapters shown per config container. Matches `specs/features/sources.feature`.

**Step 17: --source flag validation + multi-source**
Files: `cmd/unipm/install.go` (refactor)
Depends on: Step 10
Validate `--source` against available adapters. Error if unavailable with message listing available sources. Support comma-separated multi-source (`--source apt,brew`). Matches `install.feature` error + multi-source scenarios.

**Phase 2 verification**: `./unipm sources` shows adapter statuses. `./unipm install <ambiguous-pkg>` opens TUI. `./unipm uninstall <pkg>` routes correctly. `./unipm update` refreshes tracked packages. All 29 Gherkin scenarios for search/install/uninstall/update/sources pass in Tier 1 tests.

---

### Phase 3 — Adapter Expansion (target: v0.3.0)

**Step 18: Pypi adapter**
Files: `pkg/adapter/pypi.go`, `pkg/adapter/pypi_test.go`, `pkg/adapter/testdata/pypi_search_htop.json`
Depends on: Step 2
Detects `pip3` on `$PATH`. `Search()` via PyPI JSON API. `Install()` via `pip3 install --user`. `Uninstall()` via `pip3 uninstall`. Tier 1 golden fixtures + Tier 2 integration test.

**Step 19: Flatpak adapter**
Files: `pkg/adapter/flatpak.go`, `pkg/adapter/flatpak_test.go`, `pkg/adapter/testdata/flatpak_search_htop.txt`
Depends on: Step 2
Detects `flatpak` on `$PATH`. Search/install/uninstall via `flatpak` CLI. Tier 2 test calls `t.Skip()` if user namespaces unavailable per ADR-0003.

**Step 20: Brew adapter**
Files: `pkg/adapter/brew.go`, `pkg/adapter/brew_test.go`, `pkg/adapter/testdata/brew_search_htop.txt`
Depends on: Step 2
Detects `brew` on `$PATH`. All operations via `brew` CLI. Tier 1 golden fixtures + Tier 2 integration test.

**Step 21: AppImage adapter**
Files: `pkg/adapter/appimage.go`, `pkg/adapter/appimage_test.go`, `pkg/adapter/testdata/appimage_search_htop.json`
Depends on: Step 2
Detects `curl` + `wget` on `$PATH`. `Search()` via AppImageHub API. `Install()` downloads, `chmod +x`, moves to `~/Applications`. `Uninstall()` removes file. Tier 1 golden fixtures + Tier 2 integration test.

**Step 22: Register new adapters in router**
Files: `cmd/unipm/root.go` (add import + registration)
Depends on: Steps 18–21
Add pypi, flatpak, brew, appimage to the adapter registry in the Cobra root command's `init()` or startup logic.

**Step 23: Tier 2 integration test CI enabling**
Files: `.github/workflows/test.yml` (update)
Depends on: Steps 18–21
Ensure CI workflow runs `go test -tags=integration ./...` with Docker service. Add `SKIPPED_ADAPTERS` counter that warns if >2 adapters skip (allow flatpak + distrobox; alert if apt/npm/pip/brew also skip).

**Phase 3 verification**: `./unipm sources` shows 6 adapters (apt, npm, pypi, flatpak, brew, appimage). Each adapter's search/install/uninstall works. Tier 1 passes for all adapters. Tier 2 passes in CI (flatpak may skip gracefully).

---

### Phase 4 — Distrobox & Polish (target: v0.4.0)

**Step 24: Distrobox adapter**
Files: `pkg/adapter/distrobox.go`, `pkg/adapter/distrobox_test.go`
Depends on: Steps 2, 3
Reads distrobox containers from `config.yaml`. Creates one adapter per configured container (`distrobox-<name>`). `Search()` via `distrobox enter <container> -- <pm> search`. `Install()`/`Uninstall()` via distrobox enter + sudo. `IsAvailable()` checks `distrobox` on `$PATH` + container existence. Tier 2 test reserved for Tier 3 (skips in Docker per ADR-0003).

**Step 25: Register distrobox adapters**
Files: `cmd/unipm/root.go` (add logic), `cmd/unipm/sources.go` (refactor)
Depends on: Step 24
At startup, parse distrobox config and create one adapter per configured container. Register them in the router. Sources command shows them (e.g., `distrobox-arch-dev  ✓ available (yay)`).

**Step 26: Tab-completion**
Files: `cmd/unipm/completion.go`, `cmd/unipm/root.go` (add ValidArgsFunction), `cmd/unipm/search.go` (populate cache)
Depends on: All command implementations
Cobra's `GenBashCompletion`, `GenZshCompletion`, `GenFishCompletion`. `RegisterFlagCompletionFunc` for `--source` flag (completes from available adapters). `ValidArgsFunction` for package names reads from `~/.unipm/cache.json` (populated by search commands, TTL from config). No network completions for < 3 chars. Matches `specs/features/completion.feature`.

**Step 27: Documentation audit + polish**
Files: `README.md`, `CONTRIBUTING.md`, `CHANGELOG.md` (update)
Depends on: Steps 1–26
Verify README command reference matches actual behavior. Add `--version` output example. Update backend matrix with tested status per adapter. Add `CHANGELOG.md` entries for all 4 phases.

**Step 28: CI hardening**
Files: `.github/workflows/test.yml` (update), `.github/workflows/lint.yml` (update)
Depends on: Steps 1–26
Enforce coverage threshold (≥80% per adapter package). Add flaky test quarantine (3 consecutive failures → skip with issue). Ensure suite time < 5 min. Configure dependabot/renovate for Go module updates.

**Step 29: E2E nightly workflow**
Files: `.github/workflows/e2e.yml`
Depends on: Steps 24, 25
Runs full CLI workflows in Incus VM: `unipm search → install → uninstall` with all backends including distrobox and flatpak. Nightly schedule + manual dispatch. Failures open GitHub issues automatically. Does not block PR merges.

**Phase 4 verification**: `unipm completion bash | source` then `<TAB>` works. Distrobox adapter appears in `unipm sources` when configured. Nightly E2E workflow exists. All docs updated for v0.4.0.

---

## Edge Cases & Risks

### Edge cases

| # | Edge case | Handling | Gherkin scenario |
|---|-----------|----------|-----------------|
| 1 | No backends available at all | Error message listing unavailable adapters, non-zero exit | `search.feature` line 23 |
| 2 | One backend times out during search | Partial results from other backends + warning for timed-out source | `search.feature` line 29 |
| 3 | `--source` names unavailable adapter | Error listing available sources, non-zero exit | `install.feature` line 50 |
| 4 | Native installation fails | Display native error, do NOT record in state.json, non-zero exit | `install.feature` line 57 |
| 5 | Uninstall of package not in state.json | Error "was not installed via unipm" | `uninstall.feature` line 18 |
| 6 | Backend removal fails but user wants to clean state | Offer "Remove from tracking anyway? [y/N]" | `uninstall.feature` line 24 |
| 7 | Partial failure during `unipm update` (one package fails) | Report failure per-package, continue remaining, non-zero exit | `update.feature` line 30 |
| 8 | Package name collision within same source | Deduplicate by `(Source, Name)` — last result wins per invariant | `search.feature` line 13 |
| 9 | `~/.unipm/` directory does not exist on first run | Create with `0700` permissions automatically | Config/state init code |
| 10 | State file has unknown version | Error with upgrade message, never silently corrupt | `specs/architecture.md` state rules |
| 11 | Search query is empty string | Prompt user for a query string (Cobra `Args` validation) | Not in spec — defensive coding |
| 12 | Tab-completion with < 3 characters | Do not make network requests | `completion.feature` line 36 |
| 13 | Tab-completion cache expired | Do not serve stale completions; return empty until next search | `completion.feature` line 29 |

### Risks

| # | Risk | Likelihood | Impact | Mitigation |
|---|------|-----------|--------|------------|
| R1 | Adapter interface doesn't fit all backends equally (e.g., appimage Info method) | Medium | Low | Return clear "not supported" error from methods that can't be implemented; ADR-0001 acknowledges this |
| R2 | Upstream registry API changes break adapter parsing | Medium | Medium | Use JSON APIs where available (npm, PyPI); golden-file fixtures updated on every adapter change |
| R3 | Flatpak tests unreliable in CI due to user namespace issues | High | Low | `t.Skip()` with clear reason; Tier 3 nightly E2E as safety net; ADR-0003 covers this |
| R4 | Single maintainer bottleneck | High | Medium | Clear CONTRIBUTING.md; adapter interface is small (6 methods); adding adapters is isolated |
| R5 | Scope creep — users want project-local installs, source preferences | Medium | Medium | Out-of-scope list in plan.md; ADR-0002 closes source priority; revisit only if overwhelming demand |
| R6 | CI integration tests run too long (>5 min budget) | Medium | Low | Parallel adapter tests in CI; each in its own Docker container; testcontainers Ryuk cleanup |
| R7 | State file corruption on crash during write | Low | High | Atomic write (temp file + rename) eliminates this; tested explicitly |

---

## Verification

### Tier 1 — Logic tests (local, no Docker)

```bash
go test ./...                           # All packages
go test -coverprofile=coverage.out ./...  # Coverage
go tool cover -html=coverage.out         # Visual report
```

Must pass with ≥80% coverage per adapter package (`pkg/adapter/`).

### Tier 2 — Integration tests (requires Docker)

```bash
go test -tags=integration ./...          # All integration tests
go test -tags=integration -run TestApt ./pkg/adapter/  # Single adapter
```

Must pass in CI. Flatpak and distrobox may `t.Skip()` gracefully.

### Tier 3 — E2E (CI, nightly)

Full CLI workflows: `unipm search → install → uninstall → update → sources → completion`. Run in Incus VM or self-hosted runner. Does not block PR merges.

### Manual checks

- `./unipm --help` — all commands visible
- `./unipm --version` — version from `-ldflags`
- `./unipm sources` — adapters with status
- `./unipm search htop` — table with results
- `./unipm install htop --source apt` — installs, records in state.json
- `cat ~/.unipm/state.json` — record present
- `./unipm uninstall htop` — removes, cleans state.json
- `./unipm update` — updates tracked packages
- `source <(./unipm completion bash)` then `<TAB>` — completions work
