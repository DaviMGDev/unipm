# ADR-0003: Tiered testing strategy with CI offload for containerized backends

**Status**: Accepted
**Date**: 2026-07-11
**Deciders**: Davi (maintainer)

## Context

unipm's adapter suite must test 7+ heterogeneous package managers — `apt`, `npm`,
`pip`, `flatpak`, `brew`, `appimage`, and `pacman`/`dnf` inside distrobox
containers. Each backend has different system requirements, privilege models,
and containerization behavior:

- **apt, npm, pip, brew, appimage, pacman/dnf** — run cleanly in Docker
  containers via testcontainers-go.
- **flatpak** — requires systemd, D-Bus, and bubblewrap user namespaces that
  conflict with Docker's own namespace isolation. Unreliable without
  `--privileged` or custom seccomp profiles.
- **distrobox** — wraps Podman/Docker internally; nesting containers is fragile
  and not worth the maintenance cost.

Additionally, the maintainer's development machine has limited SSD space
and cannot sustain a full containerized test suite that reinstalls every
package on every run.

The `specs/stack.md` already defines three test tiers (unit, integration, E2E),
but does not specify **which adapters can run in which tiers** or **where each
tier executes** (local vs CI).

## Decision

**unipm SHALL use a three-tier testing strategy where only unit tests (Tier 1)
are mandatory locally. Full containerized integration tests (Tier 2) run on CI
only. Unreliable backends (flatpak, distrobox) are tested primarily through
logic isolation and get best-effort container tests on CI — they SHALL NOT block
local development.**

### Tier 1 — Logic tests (local, no container)

| What | Where | Blocks merge? |
|------|-------|---------------|
| Adapter output parsing (golden-file fixtures in `testdata/`) | `go test ./...` — local | Yes |
| Command flag construction (verify correct subprocess args) | `go test ./...` — local | Yes |
| Error handling (inject adapter error output, verify wrapping) | `go test ./...` — local | Yes |
| `IsAvailable()` detection via mocked `$PATH` | `go test ./...` — local | Yes |
| State read/write with `t.TempDir()` isolation | `go test ./...` — local | Yes |
| Config parsing with fixture YAML | `go test ./...` — local | Yes |
| Router dispatch with mock adapters | `go test ./...` — local | Yes |
| TUI model logic (Bubbletea model updates, no terminal) | `go test ./...` — local | Yes |

**Characteristics**: Zero containers, zero network, zero `sudo`. Total suite
should complete in under 5 seconds. This is the developer's primary feedback
loop.

### Tier 2 — Integration tests (CI, Docker containers)

| What | Where | Blocks merge? |
|------|-------|---------------|
| apt: real `apt search` / `apt install` / `apt remove` | CI (GitHub Actions) | Yes |
| npm: real `npm search` / `npm install -g` / `npm uninstall -g` | CI | Yes |
| pip: real `pip3 search` / `pip3 install --user` / `pip3 uninstall` | CI | Yes |
| brew: real `brew search` / `brew install` / `brew uninstall` | CI | Yes |
| appimage: real download + `chmod` + move + delete | CI | Yes |
| pacman/dnf: real `pacman -S` / `dnf install` in Arch/Fedora containers | CI | Yes |
| flatpak (best-effort): `flatpak search` / `flatpak install` in privileged Docker | CI | No — marked `t.Skip()` if env can't support it |
| distrobox (best-effort): `distrobox enter <container> -- <pm> search` | CI | No — skipped entirely in Docker; reserved for Tier 3 |

**Characteristics**: Containers via testcontainers-go. Run on every PR in CI.
Each adapter test runs in its own fresh container, destroyed by Ryuk after the
test. Tests `t.Skip()` gracefully when the runtime environment lacks required
capabilities (e.g., user namespaces for flatpak). Local developers MAY run
individual adapter integration tests with `go test -tags=integration -run
TestApt ./pkg/adapter/` but are not required to.

**Flatpak-specific mitigation**: Flatpak integration tests use a Docker
container with `--privileged` and
`--security-opt seccomp=unconfined`. If the CI runner's kernel configuration
kills flatpak anyway, the test calls `t.Skip("flatpak requires user namespaces
— not available in this Docker environment")` and passes. The test suite
SHALL remain green when flatpak is skipped. The adapter's logic correctness
is already covered by Tier 1 golden-file fixtures.

### Tier 3 — End-to-end (CI, VM or Incus)

| What | Where | Blocks merge? |
|------|-------|---------------|
| Full CLI workflows: `unipm search → install → uninstall` with real backends | CI (nightly or manual trigger) | No — alert only |
| Distrobox: real `unipm install --source distrobox-arch-dev` | CI (Incus VM or self-hosted runner) | No |
| Flatpak full workflow (if Tier 2 skip persists) | CI (Incus VM) | No |

**Characteristics**: Runs in an environment with a real init system (Incus
system container or KVM-backed VM on GitHub Actions). Executed nightly or on
demand via workflow dispatch. Failures generate alerts but do not block PR
merges. This tier exists to catch regressions that Tier 1 fixtures and Tier 2
best-effort containers miss — particularly flatpak sandboxing behavior and
distrobox container bridging.

### Execution matrix

| Tier | Run locally? | Run on PR CI? | Run nightly? | Storage cost (local) | Blocks merge? |
|------|-------------|---------------|-------------|----------------------|---------------|
| Tier 1 (logic) | ✅ Required | ✅ | ✅ | ~0 KB (no containers) | Yes |
| Tier 2 (integration) | ⚠️ Optional, one adapter at a time | ✅ | ✅ | ~1-3 GB per adapter (Docker images) | Yes (except skipped adapters) |
| Tier 3 (E2E) | ❌ Never | ❌ | ✅ | 0 (CI only) | No |

## Alternatives considered

| Alternative | Pros | Cons | Why rejected |
|-------------|------|------|--------------|
| All integration tests mandatory locally | Maximum correctness; no surprises on CI | Requires Docker and ~20 GB of images/containers on every dev machine; flatpak/distrobox don't work reliably in Docker anyway | Unsustainable for SSD-constrained development; doesn't work for all adapters |
| Docker Desktop / privileged Docker only for flatpak | Simpler testing setup | Requires every developer (and CI) to run Docker with `--privileged`, which is a security risk and often blocked in managed CI | The security trade-off is not justified for a CLI tool's test suite |
| Incus for everything (replace Docker entirely) | flatpak and distrobox work natively in system containers | Incus is not available on GitHub Actions managed runners; requires self-hosted runners; container spin-up is seconds instead of milliseconds | Too operationally heavy for Tier 2; better reserved for Tier 3 E2E |
| Skip flatpak and distrobox testing entirely | Zero complexity | Untested adapters rot; regressions only discovered by users | Unacceptable for a tool that advertises 7+ backends |
| Mock all adapters (never test against real PMs) | Fast, predictable, no containers ever | Doesn't catch real output format changes, subprocess argument bugs, or availability detection failures | Mock-only testing gives false confidence; Tier 2 exists precisely to validate against real package managers |

## Consequences

- **Positive**: Local development needs only `go test ./...` — sub-second
  feedback, zero disk pressure, no Docker required. CI catches integration
  regressions. The testing strategy scales to all 7+ adapters without forcing
  any single environment to support every backend. Adapter authors can add a new
  backend by writing Tier 1 fixtures first (fast iteration) and Tier 2
  containers second (CI validation).
- **Negative**: Flatpak and distrobox adapters have weaker automated coverage.
  A flatpak regression (e.g., upstream output format change) won't be caught
  until CI runs — and even then, only if the CI environment supports flatpak
  that day. The test suite has conditional skips, which can mask environment
  issues if misconfigured.
- **Mitigations**:
  - Tier 1 golden-file fixtures are **updated on every adapter change** — if
    the output format changes, the fixture update is part of the same commit.
  - Tier 2 skipped tests log a clear `t.Skip()` reason visible in CI output.
  - A `SKIPPED_ADAPTERS` counter in CI reports warns if too many adapters are
    skipping (threshold: allow flatpak + distrobox; alert if apt/npm/pip/brew
    also skip).
  - Tier 3 nightly runs serve as the safety net. If they fail, an issue is
    opened automatically and the adapter maintainer investigates before the
    next release.

## See also

- `specs/stack.md` — testing tiers and tooling (superseded in detail by this ADR)
- `specs/ops.md` — non-functional requirements (timeout handling, partial results)
- [ADR-0001](0001-adapter-pattern.md) — adapter interface definition
