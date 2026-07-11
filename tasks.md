# Tasks

## Prerequisites
- [ ] Read `specs/index.md` → confirms spec reading order
- [ ] Read all spec files in order (context → users → user_stories → features → architecture → ADRs → stack → ops)
- [ ] Confirm Go 1.22+ installed: `go version`
- [ ] Confirm Docker available (for Tier 2 tests, optional locally)

---

## Phase 1 — Skeleton

### Project Bootstrap
- [ ] **Step 1: Initialize Go module and project configs**
  - [ ] `go mod init github.com/DaviMGDev/unipm`
  - [ ] Create `.gitignore` — entries: `unipm`, `.DS_Store`, `vendor/`, `*.out`, `coverage.out`
  - [ ] Create `.golangci.yml` with gofumpt enabled, default linters
  - [ ] Create `.github/workflows/test.yml` — runs `go test ./...` and `go test -tags=integration ./...` on PR
  - [ ] Create `.github/workflows/lint.yml` — runs `golangci-lint run ./...` on PR
  - [ ] Create `.github/workflows/security.yml` — runs `gitleaks` + `govulncheck` on PR
  - **Verify:** `go build ./...` does not error (no packages yet, but module valid); CI workflow files are valid YAML (`python3 -c 'import yaml; yaml.safe_load(open(".github/workflows/test.yml"))'`)

### Core Packages (zero internal deps)
- [ ] **Step 2: Define PackageManager interface and Package struct** — traces to US-001, US-002
  - [ ] Create `pkg/adapter/adapter.go`
  - [ ] Define `Package` struct: `Name string`, `Source string`, `Version string`, `Description string`
  - [ ] Define `Details` struct (for `Info()` method return)
  - [ ] Define `PackageManager` interface: `Name()`, `Search()`, `Install()`, `Uninstall()`, `Info()`, `IsAvailable()`
  - **Verify:** `go build ./pkg/adapter` succeeds

- [ ] **Step 3: Implement config package** — traces to US-001, US-006
  - [ ] Create `pkg/config/config.go`
  - [ ] `DistroboxConfig` struct: `ContainerName string`, `PackageManager string`
  - [ ] `Config` struct: `Distrobox map[string]DistroboxConfig`, `CacheTTL int`, `SearchTimeout int`
  - [ ] `Load()` — reads `~/.unipm/config.yaml`, returns defaults if file missing
  - [ ] `Save()` — writes config with defaults to `~/.unipm/config.yaml`
  - [ ] `EnsureDir()` — creates `~/.unipm/` with `0700` permissions
  - [ ] Create `pkg/config/config_test.go` — table-driven tests with `t.TempDir()` mocked `$HOME`
  - **Verify:** `go test ./pkg/config/ -v` passes, coverage ≥80%

- [ ] **Step 4: Implement state package** — traces to US-002, US-003, US-004
  - [ ] Create `pkg/state/state.go`
  - [ ] `StateRecord` struct: `Name`, `Source`, `Version`, `InstalledAt` (RFC 3339)
  - [ ] `StateFile` struct: `Version int`, `Packages []StateRecord`
  - [ ] `Load()` — reads `~/.unipm/state.json`, validates version field
  - [ ] `Save()` — writes atomically (temp file + rename)
  - [ ] `Add(record StateRecord) error` — appends to state file if name unique
  - [ ] `Remove(name string) error` — removes record by name
  - [ ] `Get(name string) (StateRecord, error)` — finds by name
  - [ ] `List() ([]StateRecord, error)` — returns all records
  - [ ] `UpdateVersion(name, version string) error` — refreshes version field
  - [ ] Create `pkg/state/state_test.go` — atomic write tests (kill mid-write → no corruption), crud tests
  - **Verify:** `go test ./pkg/state/ -v` passes, coverage ≥80%

### First Two Adapters
- [ ] **Step 5: Implement apt adapter** — traces to US-001, US-002; `search.feature`, `install.feature`
  - [ ] Create `pkg/adapter/apt.go`
  - [ ] `AptAdapter` struct implementing `PackageManager`
  - [ ] `IsAvailable()` — `exec.LookPath("apt")` check
  - [ ] `Search()` — run `apt search <query>`, parse `name - description` lines into `[]Package`
  - [ ] `Install()` — run `sudo apt install -y <name>`
  - [ ] `Uninstall()` — run `sudo apt remove -y <name>`
  - [ ] `Info()` — run `apt show <name>`, parse output into `Details`
  - [ ] Create `pkg/adapter/apt_test.go` — `IsAvailable()` test (mocked `$PATH`), flag construction tests, error handling tests
  - [ ] Create `pkg/adapter/testdata/apt_search_htop.txt` — golden-file fixture from real `apt search htop` output
  - [ ] Parse golden fixture test: verify correct extraction of Name, Version, Description
  - **Verify:** `go test ./pkg/adapter/ -run TestApt -v` passes

- [ ] **Step 6: Implement npm adapter** — traces to US-001, US-002; `search.feature`, `install.feature`
  - [ ] Create `pkg/adapter/npm.go`
  - [ ] `NpmAdapter` struct implementing `PackageManager`
  - [ ] `IsAvailable()` — `exec.LookPath("npm")` check
  - [ ] `Search()` — HTTP GET `https://registry.npmjs.org/-/v1/search?text=<query>`, parse JSON into `[]Package`
  - [ ] `Install()` — run `npm install -g <name>`
  - [ ] `Uninstall()` — run `npm uninstall -g <name>`
  - [ ] `Info()` — run `npm info <name> --json`, parse JSON
  - [ ] Create `pkg/adapter/npm_test.go` — `IsAvailable()`, flag construction, error handling
  - [ ] Create `pkg/adapter/testdata/npm_search_htop.json` — golden-file fixture from real npm search API response
  - [ ] Parse golden fixture test
  - **Verify:** `go test ./pkg/adapter/ -run TestNpm -v` passes

### CLI Scaffold
- [ ] **Step 7: Create Cobra CLI scaffold** — traces to all stories (commands)
  - [ ] Create `cmd/unipm/main.go` — calls `rootCmd.Execute()`
  - [ ] Create `cmd/unipm/root.go`
  - [ ] Root command with `Use: "unipm"`, `Short`, `Long` (from README)
  - [ ] `--version` flag printing version injected via `-ldflags` at build time
  - [ ] Persistent `--config` flag (default `~/.unipm/config.yaml`)
  - [ ] Subcommand stubs registered: `search`, `install`, `uninstall`, `update`, `sources`, `completion`
  - [ ] Each stub prints "not yet implemented" with non-zero exit
  - [ ] `init()` loads config, checks/creates `~/.unipm/` directory
  - **Verify:** `go build -o unipm ./cmd/unipm && ./unipm --help` shows command tree with all 6 subcommands

- [ ] **Step 8: Implement search command** — traces to US-001; `search.feature` (all 5 scenarios)
  - [ ] Create `cmd/unipm/search.go`
  - [ ] `searchCmd` with `Use: "search <query>"`, `Args: cobra.ExactArgs(1)`
  - [ ] `--timeout` flag (overrides `config.SearchTimeout`)
  - [ ] Iterates all compiled-in adapters available on `$PATH`
  - [ ] Spawns goroutine per adapter, each with context timeout
  - [ ] Merges results on channel, deduplicates by `(Name, Source)`
  - [ ] Outputs table with `lipgloss`: Source | Name | Version | Description
  - [ ] Partial results + warning for timed-out adapters
  - [ ] If zero adapters available: error message, non-zero exit
  - **Verify:** `./unipm search htop` returns table with apt and/or npm results

- [ ] **Step 9: Implement install command (single-match path)** — traces to US-002; `install.feature` (single-match and --source scenarios)
  - [ ] Create `cmd/unipm/install.go`
  - [ ] `installCmd` with `Use: "install <package>"`, `Args: cobra.ExactArgs(1)`
  - [ ] `--source` / `-s` flag (string, optional)
  - [ ] If `--source` given: validate against available adapters; if unavailable → error with available list, non-zero exit
  - [ ] If `--source` given with single value: search that adapter, install if found
  - [ ] If `--source` given with comma-separated values: install from each listed source
  - [ ] If no `--source`: search all adapters; if exactly one match → install directly; if multiple → stub message "collision — TUI coming in Phase 2"
  - [ ] Install delegates to `adapter.Install(pkg)`, streams stdout/stderr
  - [ ] On success: records in `state.json` (name, source, version, installed_at)
  - [ ] On native failure: display error, do NOT record, non-zero exit
  - [ ] On success: prints "✓ <name> <version> installed from <source>"
  - **Verify:** `./unipm install htop --source apt` installs and `cat ~/.unipm/state.json` shows record

### Phase 1 Verification
- [ ] `go test ./...` passes all Tier 1 tests (adapter, state, config)
- [ ] `go build -o unipm ./cmd/unipm` produces working binary
- [ ] `./unipm --help` shows all commands
- [ ] `./unipm --version` prints version
- [ ] `./unipm search htop` returns results
- [ ] `./unipm install htop --source apt` works end-to-end
- [ ] Tag `v0.1.0` with message: "v0.1.0: initial skeleton — cobra CLI, apt + npm adapters"

---

## Phase 2 — Router & State

- [ ] **Step 10: Implement router package** — traces to US-001, US-002
  - [ ] Create `pkg/router/router.go`
  - [ ] `Router` struct: `adapters map[string]PackageManager`
  - [ ] `NewRouter() *Router`
  - [ ] `Register(adapter PackageManager)` — adds to map
  - [ ] `Get(name string) (PackageManager, error)` — retrieves by name
  - [ ] `ListAvailable() []string` — returns sorted adapter names (alphabetical per ADR-0002)
  - [ ] `SearchAll(query string, timeout time.Duration) ([]Package, []string)` — fans out to all adapters, returns results + timed-out source names
  - [ ] Create `pkg/router/router_test.go` — mock adapters implementing `PackageManager`, test fan-out, dedup, timeout, empty registry
  - **Verify:** `go test ./pkg/router/ -v` passes, coverage ≥80%

- [ ] **Step 11: Wire router into search and install commands** — traces to US-001, US-002
  - [ ] Refactor `cmd/unipm/root.go` — create `Router` at startup, register available adapters
  - [ ] Refactor `cmd/unipm/search.go` — use `router.SearchAll()` instead of direct adapter iteration
  - [ ] Refactor `cmd/unipm/install.go` — use `router.Get()` and `router.ListAvailable()` for source validation
  - **Verify:** `./unipm search htop` produces same output as before but through router dispatch

- [ ] **Step 12: Implement uninstall command** — traces to US-003; `uninstall.feature` (all 4 scenarios)
  - [ ] Create `cmd/unipm/uninstall.go`
  - [ ] `uninstallCmd` with `Use: "uninstall <package>"`, `Args: cobra.ExactArgs(1)`
  - [ ] Lookup package in `state.json` → `state.Get(name)`
  - [ ] If not found: print "`<name>` was not installed via unipm", non-zero exit
  - [ ] If found: `router.Get(record.Source)` → `adapter.Uninstall(pkg)`
  - [ ] On success: `state.Remove(name)`, print "✓ `<name>` removed from `<source>`"
  - [ ] On backend failure: print native error, offer "Remove `<name>` from unipm tracking anyway? [y/N]"
  - [ ] Read user input; if "y", `state.Remove(name)`; otherwise keep record
  - **Verify:** `./unipm uninstall <pkg>` removes package and cleans state.json

- [ ] **Step 13: Implement update command** — traces to US-004; `update.feature` (all 4 scenarios)
  - [ ] Create `cmd/unipm/update.go`
  - [ ] `updateCmd` with `Use: "update [package]"`, `Args: cobra.MaximumNArgs(1)`
  - [ ] No arg: iterate all records from `state.List()`, delegate update per adapter, refresh version
  - [ ] With arg: lookup named package, update only that one
  - [ ] If named package not in state: print "`<name>` was not installed via unipm", non-zero exit
  - [ ] Partial failure: print "✗ `<name>` (`<source>`): `<error>`" per failed package, continue remaining
  - [ ] Partial failure: non-zero exit; successfully updated packages get refreshed version
  - [ ] Each adapter's update: apt=`sudo apt upgrade <name>`, npm=`npm update -g <name>`, etc.
  - **Verify:** `./unipm update` processes all tracked packages; `./unipm update htop` updates single package

- [ ] **Step 14: Implement collision TUI** — traces to US-002; `install.feature` (collision scenario)
  - [ ] Create `pkg/ui/tui.go`
  - [ ] Bubbletea `Model`: `choices []Package`, `cursor int`, `selected *Package`, `quitting bool`
  - [ ] `Init()` returns initial command
  - [ ] `Update(msg tea.Msg)` handles: `tea.KeyMsg` (j/k/↑/↓/enter/q/esc), `tea.WindowSizeMsg`
  - [ ] `View()` renders: title, alphabetical list with cursor indicator, no preselected default
  - [ ] Sources listed alphabetically by source name (per ADR-0002)
  - [ ] `RunSelection(packages []Package) (*Package, error)` — runs the TUI program, returns selected or nil
  - [ ] Create `pkg/ui/tui_test.go` — test model state transitions (key presses update cursor, enter sets selected, esc quits)
  - **Verify:** `go test ./pkg/ui/ -v` passes

- [ ] **Step 15: Wire TUI into install command** — traces to US-002; `install.feature` (collision scenario)
  - [ ] Refactor `cmd/unipm/install.go`
  - [ ] When no `--source` and multiple adapters have the package: call `ui.RunSelection(matches)`
  - [ ] If user selects: install from that adapter, record in state.json
  - [ ] If user cancels (esc): print "installation cancelled", non-zero exit, no state record
  - **Verify:** `./unipm install requests` (where both apt and pypi have "requests") opens TUI; user selects → installs

- [ ] **Step 16: Implement sources command** — traces to US-005; `sources.feature` (all 3 scenarios)
  - [ ] Create `cmd/unipm/sources.go`
  - [ ] `sourcesCmd` with `Use: "sources"`
  - [ ] Iterate all compiled-in adapters: check `IsAvailable()`
  - [ ] Output: adapter name left-aligned, "✓ available" or "✗ not found on \$PATH"
  - [ ] Distrobox adapters: read from config, check container existence, show with `(package_manager)` suffix
  - [ ] If no adapters available: suggest installing at least one supported PM, non-zero exit
  - **Verify:** `./unipm sources` lists apt/npm with ✓, brew with ✗ (if not installed)

- [ ] **Step 17: --source flag validation + multi-source** — traces to US-002; `install.feature` (error + multi-source scenarios)
  - [ ] Refactor `cmd/unipm/install.go`
  - [ ] Parse `--source` flag: split on comma, trim whitespace
  - [ ] Validate each source name against `router.ListAvailable()`
  - [ ] If any source unavailable: error "`<source>` is not available. Available sources: `<list>`", non-zero exit
  - [ ] If multiple sources valid: install from each in order
  - [ ] Each source installation records independently in state.json
  - **Verify:** `./unipm install htop --source cargo` errors with available list; `./unipm install htop --source apt,brew` installs from both

### Phase 2 Verification
- [ ] `go test ./...` passes all Tier 1 tests
- [ ] `./unipm sources` shows adapter statuses
- [ ] `./unipm install <ambiguous-pkg>` opens TUI
- [ ] `./unipm uninstall <tracked-pkg>` routes correctly
- [ ] `./unipm uninstall <untracked-pkg>` errors with "not installed via unipm"
- [ ] `./unipm update` refreshes all tracked packages
- [ ] `./unipm update <single-pkg>` updates only that package
- [ ] `./unipm install <pkg> --source invalid` errors with available list
- [ ] `./unipm install <pkg> --source apt,brew` installs from both
- [ ] All 29 Gherkin scenarios for search/install/uninstall/update/sources traceable to passing Tier 1 tests
- [ ] Tag `v0.2.0` with message: "v0.2.0: router + state + collision TUI + uninstall + update + sources"

---

## Phase 3 — Adapter Expansion

- [ ] **Step 18: Implement pypi adapter** — traces to US-001, US-002
  - [ ] Create `pkg/adapter/pypi.go`
  - [ ] `PypiAdapter` implementing `PackageManager`
  - [ ] `IsAvailable()` — `exec.LookPath("pip3")` check
  - [ ] `Search()` — HTTP GET `https://pypi.org/pypi/<query>/json` or PyPI search API, parse JSON
  - [ ] `Install()` — run `pip3 install --user <name>`
  - [ ] `Uninstall()` — run `pip3 uninstall -y <name>`
  - [ ] `Info()` — run `pip3 show <name>`, parse output
  - [ ] Create `pkg/adapter/pypi_test.go` — golden fixture + flag construction + error handling
  - [ ] Create `pkg/adapter/testdata/pypi_search_htop.json`
  - [ ] Integration test: `//go:build integration` with testcontainers (Python container)
  - **Verify:** `go test ./pkg/adapter/ -run TestPypi -v` passes

- [ ] **Step 19: Implement flatpak adapter** — traces to US-001, US-002
  - [ ] Create `pkg/adapter/flatpak.go`
  - [ ] `FlatpakAdapter` implementing `PackageManager`
  - [ ] `IsAvailable()` — `exec.LookPath("flatpak")` check
  - [ ] `Search()` — run `flatpak search <query>`, parse columns
  - [ ] `Install()` — run `flatpak install -y <name>`
  - [ ] `Uninstall()` — run `flatpak uninstall -y <name>`
  - [ ] `Info()` — run `flatpak info <name>`, parse output
  - [ ] Create `pkg/adapter/flatpak_test.go` — golden fixture + flag construction
  - [ ] Create `pkg/adapter/testdata/flatpak_search_htop.txt`
  - [ ] Integration test: `t.Skip()` if user namespaces unavailable (per ADR-0003)
  - **Verify:** `go test ./pkg/adapter/ -run TestFlatpak -v` passes; integration test skips cleanly if env unsupported

- [ ] **Step 20: Implement brew adapter** — traces to US-001, US-002
  - [ ] Create `pkg/adapter/brew.go`
  - [ ] `BrewAdapter` implementing `PackageManager`
  - [ ] `IsAvailable()` — `exec.LookPath("brew")` check
  - [ ] `Search()` — run `brew search <query>`, parse output
  - [ ] `Install()` — run `brew install <name>`
  - [ ] `Uninstall()` — run `brew uninstall <name>`
  - [ ] `Info()` — run `brew info <name>`, parse output
  - [ ] Create `pkg/adapter/brew_test.go` — golden fixture + flag construction
  - [ ] Create `pkg/adapter/testdata/brew_search_htop.txt`
  - [ ] Integration test with testcontainers (Linuxbrew container)
  - **Verify:** `go test ./pkg/adapter/ -run TestBrew -v` passes

- [ ] **Step 21: Implement appimage adapter** — traces to US-001, US-002
  - [ ] Create `pkg/adapter/appimage.go`
  - [ ] `AppImageAdapter` implementing `PackageManager`
  - [ ] `IsAvailable()` — `exec.LookPath("curl")` OR `exec.LookPath("wget")` check
  - [ ] `Search()` — HTTP GET AppImageHub API, parse JSON into `[]Package`
  - [ ] `Install()` — download .AppImage, `chmod +x`, move to `~/Applications`
  - [ ] `Uninstall()` — remove file from `~/Applications`
  - [ ] `Info()` — extract metadata from downloaded file (return "not supported" if unavailable per ADR-0001)
  - [ ] Create `pkg/adapter/appimage_test.go` — golden fixture + flag construction
  - [ ] Create `pkg/adapter/testdata/appimage_search_htop.json`
  - [ ] Integration test with testcontainers (curl available)
  - **Verify:** `go test ./pkg/adapter/ -run TestAppImage -v` passes

- [ ] **Step 22: Register new adapters in router** — traces to US-001, US-005
  - [ ] Update `cmd/unipm/root.go` — import pypi, flatpak, brew, appimage packages
  - [ ] In startup logic: create adapter instances, check `IsAvailable()`, register if available
  - [ ] Sources command automatically picks up new adapters via `router.ListAvailable()`
  - **Verify:** `./unipm sources` shows 6 adapters (apt, npm, pypi, flatpak, brew, appimage) with availability

- [ ] **Step 23: Enable Tier 2 integration tests in CI** — traces to ADR-0003
  - [ ] Update `.github/workflows/test.yml` — add Docker service, run `go test -tags=integration ./...`
  - [ ] Add `SKIPPED_ADAPTERS` counter: parse test output for `t.Skip()`, warn if >2 adapters skipped (threshold: flatpak + distrobox)
  - [ ] If apt/npm/pip/brew/appimage also skip → CI fails (these must work in Docker)
  - **Verify:** CI run on PR shows Tier 2 tests passing or flatpak/distrobox gracefully skipped; apt/npm/pip/brew/appimage never skip

### Phase 3 Verification
- [ ] `go test ./...` passes Tier 1 for all 6 adapters
- [ ] `go test -tags=integration ./...` passes in Docker for all adapters except flatpak (may skip)
- [ ] `./unipm sources` shows all 6 adapters
- [ ] `./unipm search htop` returns results from all available backends
- [ ] Each adapter's install/uninstall works end-to-end
- [ ] Tag `v0.3.0` with message: "v0.3.0: adapter expansion — pypi, flatpak, brew, appimage"

---

## Phase 4 — Distrobox & Polish

- [ ] **Step 24: Implement distrobox adapter** — traces to US-001, US-002
  - [ ] Create `pkg/adapter/distrobox.go`
  - [ ] `DistroboxAdapter` struct: `ContainerName string`, `PackageManager string` (yay/pacman/dnf/apt)
  - [ ] Implements `PackageManager`
  - [ ] `IsAvailable()` — `exec.LookPath("distrobox")` AND `distrobox list | grep <container_name>` check
  - [ ] `Search()` — run `distrobox enter <container> -- <pm> search <query>`
  - [ ] `Install()` — run `distrobox enter <container> -- sudo <pm> -S <name>` (adapts flag per PM type)
  - [ ] `Uninstall()` — run `distrobox enter <container> -- sudo <pm> -R <name>`
  - [ ] `Name()` — returns `"distrobox-<container_name>"`
  - [ ] `Info()` — run `distrobox enter <container> -- <pm> -Qi <name>` or equivalent
  - [ ] Create `pkg/adapter/distrobox_test.go` — golden fixture + flag construction
  - [ ] Integration test: `t.Skip()` entirely in Docker (reserved for Tier 3 per ADR-0003)
  - **Verify:** `go test ./pkg/adapter/ -run TestDistrobox -v` passes Tier 1

- [ ] **Step 25: Wire distrobox adapters into startup and sources** — traces to US-005; `sources.feature` (distrobox scenario)
  - [ ] Update `cmd/unipm/root.go` — read `config.yaml` distrobox section, create one `DistroboxAdapter` per configured container
  - [ ] Register each in router: map key = `"distrobox-<container_name>"`
  - [ ] Update `cmd/unipm/sources.go` — show distrobox adapters with `(package_manager)` suffix per `sources.feature`
  - **Verify:** With distrobox config present and container running: `./unipm sources` shows `distrobox-arch-dev  ✓ available (yay)`

- [ ] **Step 26: Implement tab-completion** — traces to US-006; `completion.feature` (all 6 scenarios)
  - [ ] Create `cmd/unipm/completion.go`
  - [ ] `completionCmd` with `Use: "completion [bash|zsh|fish]"`
  - [ ] Subcommands: `completion bash`, `completion zsh`, `completion fish` using Cobra built-ins
  - [ ] On `cmd/unipm/root.go` — add `ValidArgsFunction` for `install` and `search` commands
  - [ ] `--source` flag: `RegisterFlagCompletionFunc` — completes with `router.ListAvailable()`
  - [ ] Package name completion: reads `~/.unipm/cache.json` (populated by search), filters by partial input
  - [ ] Cache TTL check: if expired, do not serve stale completions
  - [ ] If query < 3 chars: do not attempt network completions
  - [ ] `searchCmd` updated to write successful search results to `~/.unipm/cache.json`
  - **Verify:** `source <(./unipm completion bash)` then `unipm install --source <TAB>` shows available adapters; `unipm install ht<TAB>` shows cached matches

- [ ] **Step 27: Documentation audit and polish** — traces to all stories
  - [ ] Update `README.md` backend matrix — mark all 7 adapters with tested status
  - [ ] Add `--version` output example to README quickstart
  - [ ] Verify all command examples in README match actual CLI behavior
  - [ ] Add `CHANGELOG.md` entries for v0.1.0 through v0.4.0
  - [ ] Move `[Unreleased]` items to their respective version sections
  - [ ] Update `CONTRIBUTING.md` if any conventions evolved during implementation
  - **Verify:** every command in README can be copy-pasted and produces expected output

- [ ] **Step 28: CI hardening** — traces to ADR-0003
  - [ ] Update `.github/workflows/test.yml` — enforce coverage threshold (`go test -coverprofile=coverage.out && go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//'`)
  - [ ] Fail CI if `pkg/adapter/*` coverage < 80%
  - [ ] Add flaky test detection: track test results across runs, quarantine after 3 consecutive failures
  - [ ] Enforce suite time budget < 5 min in CI
  - [ ] Configure dependabot/renovate for Go module updates
  - **Verify:** CI passes with coverage check enforced

- [ ] **Step 29: Create E2E nightly workflow** — traces to ADR-0003 Tier 3
  - [ ] Create `.github/workflows/e2e.yml`
  - [ ] Schedule: `cron: '0 2 * * *'` (daily at 2 AM) + `workflow_dispatch` (manual trigger)
  - [ ] Runs in Incus VM or self-hosted runner with init system
  - [ ] Full CLI workflow: `unipm sources → search htop → install htop --source apt → verify state.json → uninstall htop → update → completion`
  - [ ] Includes distrobox workflow (config + container setup)
  - [ ] Includes flatpak workflow
  - [ ] On failure: creates GitHub issue with label `e2e-failure`
  - [ ] Does NOT block PR merges
  - **Verify:** Manual workflow dispatch succeeds; nightly schedule is active

### Phase 4 Verification
- [ ] `unipm completion bash | source` then `<TAB>` works for all completion scenarios
- [ ] Distrobox adapter appears in `unipm sources` when configured
- [ ] `unipm install <pkg> --source distrobox-arch-dev` works (if container exists)
- [ ] All 29 Gherkin scenarios have corresponding passing tests (Tier 1 or Tier 2)
- [ ] Nightly E2E workflow runs successfully
- [ ] All docs updated for v0.4.0
- [ ] Tag `v0.4.0` with message: "v0.4.0: distrobox + completion + CI polish"

---

## Verification (Final)

- [ ] **Run all Tier 1 tests** — `go test ./... -v -count=1`
- [ ] **Run all Tier 2 tests** — `go test -tags=integration ./... -v -count=1` (requires Docker)
- [ ] **Check coverage** — `go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out | grep total`
- [ ] **Run full CLI smoke test**:
  - [ ] `./unipm --help`
  - [ ] `./unipm --version`
  - [ ] `./unipm sources`
  - [ ] `./unipm search htop`
  - [ ] `./unipm install htop --source apt`
  - [ ] `cat ~/.unipm/state.json` — verify record exists
  - [ ] `./unipm uninstall htop` — verify record removed
  - [ ] `./unipm update`
  - [ ] `source <(./unipm completion bash) && unipm install --source <TAB>`
- [ ] **Manual edge case check** — verify each edge case from plan.md Edge Cases table
- [ ] **Verify all 6 user stories** — trace US-001 through US-006 to working commands
- [ ] **Verify all 29 Gherkin scenarios** — each has a passing Tier 1 or Tier 2 test
- [ ] **Update CHANGELOG.md** — create entries for each phase release

## Cleanup
- [ ] Delete `plan.md` after all tasks complete (or move to `docs/decisions/implementation-plan.md`)
- [ ] Delete `tasks.md` after all tasks complete
- [ ] Confirm completion
