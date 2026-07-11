# unipm — Architecture

## Must contain

- Entities: name, fields/types/constraints, relationships, invariants, lifecycle states
- Entity relationship overview
- Authentication method(s) and flow
- Roles, permissions, resource-level access matrix
- Token/session management
- Multi-tenancy model (if applicable)

---

## Data models

### Entity: Package

**Description**: Represents a package as returned by any backend adapter. This is the common currency of the router — all adapters produce and consume `Package` values.

**Fields**:

| Field | Type | Required | Unique | Default | Constraints |
|-------|------|----------|--------|---------|-------------|
| `name` | `string` | Yes | No (per-source) | — | Non-empty |
| `source` | `string` | Yes | No (per-source) | — | Must match a registered adapter name |
| `version` | `string` | No | No | `""` | Free-form; backend-specific format |
| `description` | `string` | No | No | `""` | May be truncated from backend output |

**Relationships**:

- `Package` → `StateRecord`: A `Package` that has been installed becomes a `StateRecord` with an additional `installed_at` timestamp.

**Invariants**:

- `(source, name)` uniquely identifies a package within a search result set. No two results in the same output table may share both fields.

### Entity: StateRecord

**Description**: A record in `~/.unipm/state.json` tracking a package that was installed through unipm.

**Fields**:

| Field | Type | Required | Unique | Default | Constraints |
|-------|------|----------|--------|---------|-------------|
| `name` | `string` | Yes | Yes (within the array) | — | Non-empty |
| `source` | `string` | Yes | No | — | Must match a registered adapter name |
| `version` | `string` | Yes | No | — | Snapshot at install/update time |
| `installed_at` | `string` (RFC 3339) | Yes | No | — | UTC timestamp |

**Relationships**:

- `StateRecord` → `PackageManager` adapter: The `source` field names the adapter that installed (and can uninstall/update) this package.

**Lifecycle states**:

```
[Not tracked] → [Tracked] → [Removed]
```

- Transition to **Tracked** triggered by: successful `unipm install`.
- Transition to **Removed** triggered by: successful `unipm uninstall`.
- On `unipm update`, the record stays **Tracked** with an updated `version`.

**Invariants**:

- A package name must be unique within the state file array. Installing the same package name from two different sources is not tracked as a single record — each `(source, name)` pair is distinct.
- The state file version field (`"version": 1`) must be validated before reading records.

### Entity: PackageManager (Interface)

**Description**: The Go interface every backend adapter implements. The router holds a registry of adapters and calls them polymorphically.

**Fields (methods)**:

| Method | Signature | Description |
|--------|-----------|-------------|
| `Name()` | `() string` | Returns the adapter identifier (e.g., `"apt"`, `"npm"`) |
| `Search(query string)` | `([]Package, error)` | Queries the backend for packages matching `query` |
| `Install(pkg Package)` | `error` | Delegates installation to the native package manager |
| `Uninstall(pkg Package)` | `error` | Delegates removal to the native package manager |
| `Info(pkg Package)` | `(Details, error)` | Returns extended metadata for a package |
| `IsAvailable()` | `bool` | Checks whether the required binary is on `$PATH` |

**Invariants**:

- `IsAvailable()` must be called before `Search()`, `Install()`, or `Uninstall()` — the router enforces this at startup by building the adapter registry from available adapters only.
- `Install()` and `Uninstall()` receive a `Package` whose `source` matches the adapter's `Name()`.

### Entity: DistroboxConfig

**Description**: A configured distrobox container defined in `~/.unipm/config.yaml`.

**Fields**:

| Field | Type | Required | Unique | Default | Constraints |
|-------|------|----------|--------|---------|-------------|
| `container_name` | `string` | Yes | Yes | — | Must match an existing distrobox container |
| `package_manager` | `string` | Yes | No | — | One of: `apt`, `pacman`, `yay`, `dnf`, `zypper` |

**Relationships**:

- `DistroboxConfig` → `PackageManager`: Each config entry generates a `distrobox-<container_name>` adapter at startup.

### Entity relationship diagram

```
┌─────────────────┐     implements      ┌─────────────────────┐
│  PackageManager  │◄────────────────────│  apt / npm / pypi /  │
│   (interface)    │                     │  flatpak / brew /    │
│                  │                     │  appimage / distrobox│
└────────┬────────┘                     └─────────────────────┘
         │ produces / consumes
         ▼
┌─────────────────┐     installed as     ┌─────────────────────┐
│    Package       │────────────────────►│    StateRecord       │
│                  │                     │  (in state.json)     │
│ - name           │                     │                      │
│ - source         │                     │ - name               │
│ - version        │                     │ - source             │
│ - description    │                     │ - version            │
└─────────────────┘                     │ - installed_at       │
                                         └─────────────────────┘

┌──────────────────────┐
│  DistroboxConfig      │
│  (in config.yaml)     │
│                       │
│ - container_name      │
│ - package_manager     │
└──────────────────────┘
```

---

## Concurrency model

- Network calls to registries (npm, PyPI, etc.) are issued concurrently via goroutines with a configurable timeout (default 10 s).
- Shell-out commands (`apt search`, `flatpak search`, etc.) are run in parallel where the backend supports it.
- Results are merged on a channel and deduplicated by `(Source, Name)`.
- On startup, the router iterates all compiled-in adapters and calls `IsAvailable()` on each — available adapters are registered; unavailable ones are skipped.

---

## Backend adapter contracts

Every adapter implements the `PackageManager` interface. No adapter is classified as "system", "language", "containerised", or any other category — all are equal from unipm's perspective.

### apt

| Property | Behaviour |
|----------|-----------|
| Detection | Check for `apt` on `$PATH` |
| Search | `apt search` |
| Install | `sudo apt install` (elevates only this subprocess) |
| Uninstall | `sudo apt remove` |
| Info | `apt show` |

### pacman / dnf

| Property | Behaviour |
|----------|-----------|
| Detection | Check for binary on `$PATH` |
| Search | `pacman -Ss` / `dnf search` |
| Install | `sudo pacman -S` / `sudo dnf install` |
| Uninstall | `sudo pacman -R` / `sudo dnf remove` |
| Info | `pacman -Qi` / `dnf info` |

### npm

| Property | Behaviour |
|----------|-----------|
| Detection | Check for `npm` on `$PATH` |
| Search | `npm search` (or registry API) |
| Install | `npm install -g` (global scope by default) |
| Uninstall | `npm uninstall -g` |
| Info | `npm info` |

### pypi

| Property | Behaviour |
|----------|-----------|
| Detection | Check for `pip3` on `$PATH` |
| Search | `pip3 search` (or PyPI JSON API) |
| Install | `pip3 install --user` (user-wide by default) |
| Uninstall | `pip3 uninstall` |
| Info | `pip3 show` |

### flatpak

| Property | Behaviour |
|----------|-----------|
| Detection | Check for `flatpak` on `$PATH` |
| Search | `flatpak search` |
| Install | `flatpak install` |
| Uninstall | `flatpak uninstall` |
| Info | `flatpak info` |

### appimage

| Property | Behaviour |
|----------|-----------|
| Detection | Check for `curl` + `wget` on `$PATH` |
| Search | Query AppImageHub API / GitHub releases |
| Install | Download `.AppImage`, `chmod +x`, move to `~/Applications` |
| Uninstall | Remove file from `~/Applications` (no upstream registry) |
| Info | Extract metadata from the AppImage file |

### brew

| Property | Behaviour |
|----------|-----------|
| Detection | Check for `brew` on `$PATH` |
| Search | `brew search` |
| Install | `brew install` |
| Uninstall | `brew uninstall` |
| Info | `brew info` |

### distrobox-* (one per configured container)

| Property | Behaviour |
|----------|-----------|
| Detection | Check for `distrobox` on `$PATH` + container existence |
| Search | `distrobox enter <container> -- <pm> search` |
| Install | `distrobox enter <container> -- sudo <pm> -S` |
| Uninstall | `distrobox enter <container> -- sudo <pm> -R` |
| Config | Defined in `~/.unipm/config.yaml` (container name, PM) |

---

## State model

### Location

`~/.unipm/state.json`

### Schema

```json
{
  "version": 1,
  "packages": [
    {
      "name": "htop",
      "source": "apt",
      "version": "3.3.0",
      "installed_at": "2026-07-10T12:00:00Z"
    }
  ]
}
```

### Rules

- Only packages installed via `unipm install` are recorded.
- On `unipm uninstall`, the record is removed after successful removal.
- On `unipm update`, the `version` field is refreshed.
- The state file is **not** a lock file — the user can still use native package managers directly; unipm simply won't track packages installed outside itself.

---

## Configuration schema

### Location

`~/.unipm/config.yaml`

### Schema

```yaml
# Distrobox container definitions
distrobox:
  arch:
    container_name: arch-dev
    package_manager: yay        # or pacman
  fedora:
    container_name: fedora-toolbox
    package_manager: dnf

# Cache TTL for tab-completion (seconds)
cache_ttl: 86400

# Default search timeout per adapter (seconds)
search_timeout: 10
```

**Note:** There is no priority or preference setting. unipm does not rank sources. The collision TUI lists all matching sources alphabetically; the user chooses.

---

## Authentication & authorization

**N/A** — unipm is a local CLI tool. There are no user accounts, login flows, or role-based access controls. Privilege escalation for system-level package operations is handled by scoping `sudo` to the backend subprocess only.

## Multi-tenancy

**N/A** — unipm runs as a single-user local binary. Configuration and state files live in the user's home directory (`~/.unipm/`).

## See also

- `adr/` — architectural decisions (incl. adapter pattern rationale and no-priority design)
- `stack.md` — technology that implements this architecture
- `ops.md` — distribution, config paths, and roadmap
