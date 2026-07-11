# unipm — Universal Package Manager

**unipm** is a meta package manager that unifies multiple package ecosystems under
a single command-line interface. Search, install, and remove software from `apt`,
`npm`, `pypi`, `flatpak`, `AppImage`, `brew`, and even `pacman`/`yay` inside a
Distrobox container — without leaving your terminal or memorising a dozen
different flags.

```
unipm install requests
# Found 'requests' in 3 sources. Select one:
#   > [pypi]  requests (2.31.0) - Python HTTP for Humans
#     [npm]   requests (1.0.4)  - Simple HTTP request library
#     [apt]   python3-requests  - HTTP library for Python
```

---

## Features

- **Unified search** — query all registries in parallel with one command.
- **Smart collision handling** — when a package exists in multiple sources,
  unipm shows an interactive TUI that lets you pick.
- **Explicit source flag** — bypass the prompt with `--source apt` or `-s pypi`.
- **State tracking** — remembers where each package was installed so
  `unipm uninstall` and `unipm update` just work.
- **Tab completion** — for `--source` values and even package names (with local
  caching).
- **Zero lock-in** — every action delegates to the native package manager; your
  system stays consistent.

---

## Installation

### From source (Go required)

```bash
git clone https://github.com/DaviMGDev/unipm.git
cd unipm
go build -o unipm ./cmd/unipm
sudo cp unipm /usr/local/bin/
```

### From release binaries

Download the latest binary for your platform from the
[releases page](https://github.com/DaviMGDev/unipm/releases), make it executable,
and place it in your `$PATH`:

```bash
chmod +x unipm
sudo mv unipm /usr/local/bin/
```

### Enable tab completion

Add one of the following to your shell config file:

```bash
# Bash
eval "$(unipm completion bash)"

# Zsh
eval "$(unipm completion zsh)"

# Fish
unipm completion fish | source
```

---

## Quickstart

1. **Check which sources are available on your system:**

   ```bash
   unipm sources
   ```

2. **Search for a package across all sources:**

   ```bash
   unipm search htop
   ```

3. **Install a package** (interactively pick the source if there's a collision):

   ```bash
   unipm install htop
   ```

4. **Install from a specific source** (no prompt):

   ```bash
   unipm install httpie --source apt
   ```

5. **Uninstall** (unipm remembers where it was installed from):

   ```bash
   unipm uninstall htop
   ```

6. **Update everything installed via unipm:**

   ```bash
   unipm update
   ```

---

## Command Reference

### `unipm search <query>`

Query all available backends and display results in a unified table.

| Flag | Description |
|------|-------------|
| (none) | Searches all sources in parallel |

**Example:**

```bash
unipm search jq
# Source  Name    Version  Description
# apt     jq      1.6      Command-line JSON processor
# brew    jq      1.6      Lightweight JSON processor
```

### `unipm install <package> [--source <source>]`

Install a package. If no `--source` is given and the package exists in multiple
sources, an interactive TUI opens for selection.

| Flag | Short | Description |
|------|-------|-------------|
| `--source` | `-s` | Source to install from (e.g., `apt`, `npm`, `pypi`) |

**Examples:**

```bash
# Interactive (collision prompt if ambiguous)
unipm install requests

# Explicit source
unipm install requests --source pypi

# Short flag
unipm install requests -s apt

# Install from multiple sources
unipm install requests -s pypi,npm
```

### `unipm uninstall <package>`

Uninstall a package that was installed via unipm. Looks up the source from the
local state file.

```bash
unipm uninstall htop
```

### `unipm update [package]`

Update packages managed by unipm.

- Without an argument: update all tracked packages.
- With a package name: update only that package.

```bash
# Update everything
unipm update

# Update a single package
unipm update htop
```

### `unipm sources`

List all configured package sources and their availability status.

```bash
unipm sources
# apt       ✓
# npm       ✓
# pypi      ✓
# flatpak   ✓
# appimage  ✓
# brew      ✗ (not on $PATH)
# arch      ✓
```

### `unipm completion [bash|zsh|fish]`

Generate shell completion script.

```bash
unipm completion zsh > /usr/local/share/zsh/site-functions/_unipm
```

---

## Configuration

unipm stores its configuration at `~/.unipm/config.yaml`. The file is created
automatically on first run with sensible defaults.

Key settings:

- **`distrobox`** — Container definitions for Distrobox-managed environments.
- **`cache_ttl`** — How long search results are cached for tab completion.
- **`search_timeout`** — Per-adapter timeout for search queries.

Example:

```yaml
distrobox:
  arch:
    container_name: arch-dev
    package_manager: yay

cache_ttl: 86400
search_timeout: 10
```

**Note:** There is no priority or preference setting. unipm does not rank
sources. The collision TUI lists all matching sources alphabetically; the user
chooses.

---

## Backend Matrix

| Source     | Search | Install | Uninstall | Detection              |
|------------|--------|---------|-----------|------------------------|
| apt        | ✓      | ✓ (sudo)| ✓ (sudo)  | `apt` on `$PATH`       |
| npm        | ✓      | ✓ ( -g )| ✓ ( -g )  | `npm` on `$PATH`       |
| pypi       | ✓      | ✓ (--user)| ✓        | `pip3` on `$PATH`      |
| flatpak    | ✓      | ✓       | ✓         | `flatpak` on `$PATH`   |
| appimage   | ✓      | ✓       | ✓¹        | `curl` on `$PATH`      |
| brew       | ✓      | ✓       | ✓         | `brew` on `$PATH`      |
| distrobox  | ✓      | ✓ (sudo)| ✓ (sudo)  | `distrobox` + container|

¹ AppImage uninstall removes the file from `~/Applications`; tracking depends on
unipm's state file.

---

## Local State

unipm maintains a state database at `~/.unipm/state.json` that records every
package installed through the tool:

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

This file is the source of truth for `unipm uninstall` and `unipm update`.
Packages installed outside unipm (e.g., manually via `apt`) are not tracked.

---

## Contributing

Contributions are welcome! Here's how to get started:

1. Fork the repository.
2. Create a feature branch (`git checkout -b feat/my-feature`).
3. Commit your changes with conventional commit messages.
4. Push to your fork and open a pull request.

### Development requirements

- Go 1.22+
- `golangci-lint` (for linting)
- The native package manager you are developing an adapter for

### Project structure

```
cmd/unipm/           # Main binary entrypoint
pkg/adapter/         # PackageManager interface + all adapters
pkg/router/          # Adapter registry and fan-out logic
pkg/state/           # state.json read/write
pkg/ui/              # Bubbletea TUI components
pkg/config/          # Config file parsing
```

### Coding conventions

- Follow `gofumpt` formatting.
- Write tests for every adapter using the host's actual package manager
  (CI uses containers for isolation).
- Use conventional commits (`feat:`, `fix:`, `docs:`, `refactor:`, etc.).

---

## License

unipm is released under the MIT License. See [LICENSE](./LICENSE) for details.
