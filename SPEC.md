# unipm — Universal Package Manager Specification

**Version:** 0.1.0-draft  
**Status:** Design  
**Language:** Go (Golang)  
**License:** MIT  

---

## 1. Overview

unipm is a meta-package-manager CLI that aggregates multiple heterogeneous package
ecosystems under a single command-line interface. It acts as a **router**: when the
user searches for a package, unipm queries all available backends in parallel and
presents a unified result set. When the user installs a package, unipm either
routes directly to the sole matching source or, when a collision exists, presents
an interactive TUI prompt that lets the user choose.

### Goals

- Eliminate context-switching between `apt`, `npm`, `pip`, `flatpak`, etc.
- Provide a single, discoverable command surface (`search`, `install`, `uninstall`,
  `update`, `sources`).
- Respect each backend's native behaviour while
  offering a consistent UX.
- Track only packages installed *through* unipm so the local system state file
  stays lightweight and authoritative for `unipm uninstall` / `unipm update`.

### Non-Goals

- Replacing the upstream package managers or their metadata.
- Managing project-local dependencies (no `--dev` / `--save` flags — default to
  user-wide/global installs).
- Running the entire tool as root; privilege escalation is scoped to the backend
  subprocess that requires it.
- Ranking, categorising, or expressing preference for any package source.
  unipm has no opinion on which source is "system", "language", or "containerised".
  It is a pure router.

---

## 2. Architecture

### 2.1 Pattern: Adapter (Plugin)

Every backend implements a common `PackageManager` interface. The main binary (The
Router) holds a registry of adapters and calls them polymorphically.

```go
type PackageManager interface {
    Name() string                      // e.g., "apt", "npm", "distrobox-arch"
    Search(query string) ([]Package, error)
    Install(pkg Package) error
    Uninstall(pkg Package) error
    Info(pkg Package) (Details, error)
    IsAvailable() bool                 // checks whether the tool is on $PATH
}
```

### 2.2 Router

- On startup, the router iterates all compiled-in adapters and calls
  `IsAvailable()` on each.
- A command like `search` fans out goroutines to every available adapter,
  merges the results, and displays them in a colourised table.
- The `--source` flag pre-selects a specific adapter, bypassing the fan-out /
  collision logic.

### 2.3 Concurrency Model

- Network calls to registries (npm, PyPI, etc.) are issued concurrently via
  goroutines with a configurable timeout (default 10 s).
- Shell-out commands (`apt search`, `flatpak search`, etc.) are run in parallel
  where the backend supports it.
- Results are merged on a channel and deduplicated by `(Source, Name)`.

### 2.4 TUI Layer

- Built with **Bubbletea** (Charmbracelet).
- Used only for the collision-resolution prompt on `install`.
- Separate from the main CLI output (plain table for `search`, TUI for ambiguous
  `install`).
- **No preselected default.** Items are listed in the order returned by the
  router (alphabetical by source name). unipm does not express a preference.

---

## 3. CLI Reference

### 3.1 Global Flags

| Flag            | Short | Description                          |
|-----------------|-------|--------------------------------------|
| `--config`      |       | Path to config file (default `~/.unipm/config.yaml`) |
| `--debug`       |       | Enable debug logging                 |
| `--help`        | `-h`  | Show help                            |

### 3.2 Commands

#### `unipm search <query>`

- Queries **all** available adapters in parallel.
- Returns a unified table of `Source | Name | Version | Description`.
- No side effects — never modifies system state.

#### `unipm install <package> [--source <source>]`

- **No `--source`:** Searches all adapters. If exactly one match → install
  directly. If multiple matches → open interactive TUI for selection. If zero
  matches → error with suggestion.
- **`--source <source>:`** Installs directly from the named adapter. Errors if
  the source is unavailable or the package is not found there.
- **`--source <src1,src2>:`** Installs the package from every listed source
  (useful when the same logical package exists in, e.g., PyPI and npm for
  different ecosystems).

#### `unipm uninstall <package>`

- Looks up the package in the local state database (`~/.unipm/state.json`).
- Routes the uninstall command to the adapter that originally installed it.
- Errors if the package is not tracked by unipm (i.e., was installed manually).

#### `unipm update [package]`

- **Without argument:** checks all packages tracked in the local state for
  newer versions and upgrades them.
- **With argument:** updates only the named package (must be tracked).
- Delegates to each backend's native update mechanism.

#### `unipm sources`

- Lists every adapter with its status:
  ```
  apt       ✓ available
  npm       ✓ available
  pypi      ✓ available
  flatpak   ✓ available
  appimage  ✓ available
  brew      ✗ not found on $PATH
  arch      ✓ available
  ```

#### `unipm completion [bash|zsh|fish]`

- Generates shell-completion script (built-in Cobra feature).
- User pipes the output into the appropriate completions directory.

---

## 4. Backend Adapter Contracts

Every adapter implements the same `PackageManager` interface. The table below
describes the specific behaviour of each backend. No adapter is classified as
"system", "language", "containerised", or any other category — all are equal
from unipm's perspective.

### 4.1 `apt`

| Property     | Behaviour                                              |
|--------------|--------------------------------------------------------|
| Detection    | Check for `apt` on `$PATH`                             |
| Search       | `apt search`                                           |
| Install      | `sudo apt install` (elevates only this subprocess)     |
| Uninstall    | `sudo apt remove`                                      |
| Info         | `apt show`                                             |

### 4.2 `pacman` / `dnf`

| Property     | Behaviour                                              |
|--------------|--------------------------------------------------------|
| Detection    | Check for binary on `$PATH`                            |
| Search       | `pacman -Ss` / `dnf search`                            |
| Install      | `sudo pacman -S` / `sudo dnf install`                  |
| Uninstall    | `sudo pacman -R` / `sudo dnf remove`                   |
| Info         | `pacman -Qi` / `dnf info`                              |

### 4.3 `npm`

| Property     | Behaviour                                              |
|--------------|--------------------------------------------------------|
| Detection    | Check for `npm` on `$PATH`                             |
| Search       | `npm search` (or registry API)                         |
| Install      | `npm install -g` (global scope by default)             |
| Uninstall    | `npm uninstall -g`                                     |
| Info         | `npm info`                                             |

### 4.4 `pypi`

| Property     | Behaviour                                              |
|--------------|--------------------------------------------------------|
| Detection    | Check for `pip3` on `$PATH`                            |
| Search       | `pip3 search` (or PyPI JSON API)                       |
| Install      | `pip3 install --user` (user-wide by default)           |
| Uninstall    | `pip3 uninstall`                                       |
| Info         | `pip3 show`                                            |

### 4.5 `flatpak`

| Property     | Behaviour                                              |
|--------------|--------------------------------------------------------|
| Detection    | Check for `flatpak` on `$PATH`                         |
| Search       | `flatpak search`                                       |
| Install      | `flatpak install`                                      |
| Uninstall    | `flatpak uninstall`                                    |
| Info         | `flatpak info`                                         |

### 4.6 `appimage`

| Property     | Behaviour                                              |
|--------------|--------------------------------------------------------|
| Detection    | Check for `curl` + `wget` on `$PATH`                   |
| Search       | Query AppImageHub API / GitHub releases                |
| Install      | Download `.AppImage`, `chmod +x`, move to `~/Applications` |
| Uninstall    | Remove file from `~/Applications` (no upstream registry)|
| Info         | Extract metadata from the AppImage file                |

### 4.7 `brew`

| Property     | Behaviour                                              |
|--------------|--------------------------------------------------------|
| Detection    | Check for `brew` on `$PATH`                            |
| Search       | `brew search`                                          |
| Install      | `brew install`                                         |
| Uninstall    | `brew uninstall`                                       |
| Info         | `brew info`                                            |

### 4.8 `distrobox-*` (one per configured container)

| Property     | Behaviour                                              |
|--------------|--------------------------------------------------------|
| Detection    | Check for `distrobox` on `$PATH` + container existence |
| Search       | `distrobox enter <container> -- <pm> search`           |
| Install      | `distrobox enter <container> -- sudo <pm> -S`          |
| Uninstall    | `distrobox enter <container> -- sudo <pm> -R`          |
| Config       | Defined in `~/.unipm/config.yaml` (container name, PM) |

---

## 5. State Model

### 5.1 Location

`~/.unipm/state.json`

### 5.2 Schema

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

### 5.3 Rules

- Only packages installed via `unipm install` are recorded.
- On `unipm uninstall`, the record is removed after successful removal.
- On `unipm update`, the `version` field is refreshed.
- The state file is **not** a lock file — the user can still use native package
  managers directly; unipm simply won't track packages installed outside itself.

---

## 6. Configuration Schema

### 6.1 Location

`~/.unipm/config.yaml`

### 6.2 Schema

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

**Note:** There is no priority or preference setting. unipm does not rank
sources. The collision TUI lists all matching sources alphabetically; the user
chooses.

---

## 7. Tab-Completion Design

### 7.1 Mechanism

Cobra's built-in `RegisterFlagCompletionFunc` for `--source` values and
`ValidArgsFunction` for package-name arguments.

### 7.2 Source Values

The completion callback returns the list of adapters where
`IsAvailable() == true`, filtered by the user's partial input.

### 7.3 Package Names

A local cache at `~/.unipm/cache.json` (populated by previous `search`
commands) serves completions instantly. The cache has a configurable TTL
(default 24 h). For queries < 3 characters, no network-backed completions
are attempted.

---

## 8. Implementation Roadmap

### Phase 1 — Skeleton

- Go module init, Cobra CLI scaffold.
- `PackageManager` interface in `pkg/adapter/`.
- Two adapters: `apt` and `npm`.
- `search` and `install` working for both.
- Basic result table output.

### Phase 2 — Router & State

- `state.json` read/write in `~/.unipm/`.
- `uninstall` and `update` commands.
- Collision TUI (Bubbletea) for ambiguous `install` — no preselected default,
  items listed alphabetically by source name.
- `sources` command.
- `--source` flag validation against available adapters.

### Phase 3 — Adapter Expansion

- Adapters: `pypi`, `flatpak`, `brew`.
- `sudo` elevation scoped to system-package subprocesses.
- `appimage` adapter (AppImageHub API, download, `chmod`, move).

### Phase 4 — Distrobox & Polish

- `distrobox` adapter with config-file parsing.
- Tab-completion: `RegisterFlagCompletionFunc` + `ValidArgsFunction` +
  local cache.
- Documentation, test suite, CI.
