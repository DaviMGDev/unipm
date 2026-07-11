# unipm — Stack

## Must contain

- Technology: category, name, pinned version, purpose, rationale, constraints
- External services: provider, integration method, criticality, fallback
- Testing: tiers, coverage targets, test data policy, CI rules

---

## Languages & frameworks

| Name | Version | Used for | Rationale | Constraints |
|------|---------|----------|-----------|-------------|
| Go | 1.22+ | Entire codebase | Single-binary compilation, strong stdlib (subprocess management, concurrency), cross-platform builds, rich CLI ecosystem | Must use `gofumpt` formatting |
| Cobra | 1.x | CLI framework | De-facto standard for Go CLIs; built-in help, completion, flag parsing; used by kubectl, helm, hugo | — |
| Bubbletea | 1.x | TUI for collision prompt | Elm-architecture TUI framework from Charmbracelet; lightweight, terminal-native, mouse/keyboard support | Used only for the collision-resolution prompt; not the main CLI output |

## Databases & infrastructure tools

**N/A** — unipm is a local CLI tool with no database. State is stored in a single JSON file (`~/.unipm/state.json`). Configuration is stored in a YAML file (`~/.unipm/config.yaml`).

## Development tooling

| Tool | Version | Used for | Rationale |
|------|---------|----------|-----------|
| gofumpt | latest | Code formatting | Stricter `gofmt`; enforced in CI |
| golangci-lint | latest | Linting | Aggregates multiple Go linters; configured for CI |
| Go test | built-in | Unit + integration testing | stdlib testing with table-driven tests |
| testcontainers-go | latest | Integration test isolation | Spins up containers with actual package managers for adapter testing |

## External services

**None** — unipm is fully local. All operations delegate to locally installed package manager binaries (`apt`, `npm`, `pip3`, `flatpak`, etc.). Network calls (e.g., to PyPI JSON API, npm registry, AppImageHub) are made directly by the adapters — there are no third-party API dependencies or external services that unipm itself depends on.

### External service risk

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Upstream registry outage (npm, PyPI) | Low | Low — affects only that backend's search/install | Show partial results from other backends; display clear timeout error for the affected source |
| Upstream registry API change | Low | Medium — adapter parsing may break | Adapters parse structured output (JSON APIs) where available; fall back to CLI output parsing with version-pinned test fixtures |

### Compliance

- No data is sent to third parties by unipm itself. Adapters make network requests directly to package registries (npm, PyPI, AppImageHub) when the user initiates a search or install.
- No user PII is collected, stored, or transmitted.

## Testing

### Tiers

| Tier | Scope | Tools | Runs on | Blocks merge? |
|------|-------|-------|---------|---------------|
| **Unit** | Individual adapters, state read/write, config parsing, router logic | `go test` | Every commit / PR | Yes |
| **Integration** | Adapter-to-real-PM interaction (apt installs in a container, npm searches against registry) | `go test` + testcontainers | Every PR | Yes |
| **E2E** | Full CLI workflows: `unipm search → install → uninstall` with real backends | Shell scripts + containers | PR + nightly | No — alert only |
| **Security** | Dependency scanning, secret detection | gitleaks, govulncheck | Every PR | Yes |

### Coverage targets

| Tier | Metric | Target | Enforcement |
|------|--------|--------|-------------|
| Unit | Line coverage | 80% per adapter package | CI fails below threshold |
| Integration | Critical path coverage | Every adapter has at least one happy-path install/uninstall test | CI warns |

### Test data policy

- **Fixtures**: Static JSON/YAML fixtures for state and config parsing tests. Use golden-file patterns for adapter output parsing.
- **Seeding**: Integration tests use testcontainers with fresh container images — no shared state.
- **Isolation**: Tests must not depend on execution order. Each test creates its own temp directory for `state.json` and `config.yaml`.
- **Secrets**: No real credentials. Test-only configuration uses mocked `$HOME` paths.

### File conventions

| Concern | Convention | Example |
|---------|------------|---------|
| Location | Co-located `_test.go` files | `pkg/adapter/apt_test.go` |
| Naming | `*_test.go` | `pkg/state/state_test.go` |

### CI rules

1. All tests in tiers marked "Blocks merge" must pass before merge.
2. Flaky test policy: quarantine after 3 consecutive failures on main.
3. Suite time budget: < 5 minutes.
4. Coverage reports generated on every PR; line-coverage threshold enforced.

## See also

- `architecture.md` — what this stack implements
- `adr/` — rationale for major technology choices (adapter pattern, no priority)
- `ops.md` — where this stack runs (distribution, config paths)
- `CONTRIBUTING.md` — conventions (commits, PRs, naming)
