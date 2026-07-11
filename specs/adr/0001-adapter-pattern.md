# ADR-0001: Adapter pattern for backend dispatch

**Status**: Accepted
**Date**: 2026-07-10
**Deciders**: Davi (maintainer)

## Context

unipm must support 7+ heterogeneous package managers (`apt`, `npm`, `pip`, `flatpak`, `brew`, `pacman`, `distrobox`), each with different search syntaxes, install flags, privilege models, and output formats. The dispatch mechanism must allow adding new backends without modifying the router core, and each backend must be testable in isolation.

The alternative is a monolithic `switch`/`if-else` dispatch inside the router, where each command handler inspects the source string and branches into backend-specific logic. This requires touching the router every time a backend is added, changed, or removed.

## Decision

Use the **Adapter (Plugin) pattern**: every backend implements a common `PackageManager` Go interface. The router holds a registry of adapters (a `map[string]PackageManager`) populated at startup by iterating all compiled-in adapters and calling `IsAvailable()` on each. All commands (`search`, `install`, `uninstall`, `update`) call the interface methods polymorphically.

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

## Alternatives considered

| Alternative | Pros | Cons | Why rejected |
|-------------|------|------|--------------|
| Monolithic switch dispatch | Simpler for 2–3 backends; no interface abstraction overhead | Router becomes a growing tangle of backend-specific logic; adding a backend requires modifying the router; testing requires mocking the entire router | Does not scale to 7+ backends with distinct behaviors |
| External plugin system (e.g., HashiCorp go-plugin) | Backends can be distributed separately; hot-reloadable | Adds RPC serialization overhead; complex deployment (plugin binaries must match host architecture); overkill for a statically-compiled CLI | unipm is distributed as a single binary; compiled-in adapters are sufficient and simpler |
| Shell-script wrappers (each backend is a shell script) | Zero compilation; easy for users to add custom backends | No type safety; error handling is string-based; testing is fragile; performance penalty from subprocess-per-command | Go's type system and testing tooling are major advantages for correctness |

## Consequences

- **Positive**: Adding a new backend means writing one file that implements `PackageManager` and registering it in the adapter list — the router never changes. Each adapter can be tested in isolation with a mock package manager binary. The interface contract is small (6 methods) and easy to understand.
- **Negative**: The interface forces all backends into the same method signatures, which may not perfectly fit every backend's native behavior. For example, `appimage` has no upstream registry for `Info()` — it must extract metadata from the downloaded file, which is a different operation than `apt show`. The adapter must handle these mismatches internally.
- **Mitigations**: Backends that cannot fully implement a method return a clear "not supported" error rather than silently failing. The `IsAvailable()` gate prevents unavailable backends from being called at all.
