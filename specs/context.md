# unipm — Context

## Must contain

- One-paragraph problem statement
- Target audience (high level)
- Domain boundaries: in scope, out of scope
- Goals with measurable success metrics
- Non-goals with rationale
- Stakeholders: role, interest, influence, key concerns
- Key assumptions
- Project phase

---

## Problem statement

Developers and system administrators juggle a dozen package managers — `apt`, `npm`, `pip`, `flatpak`, `brew`, `pacman`, and others — each with its own syntax, flags, and mental model. Searching for a package means guessing which ecosystem it lives in, and uninstalling means remembering where it came from. unipm eliminates this context-switching by providing a single, consistent CLI that routes commands to the correct backend automatically.

## Target audience

- **System administrators** — who install and maintain user-facing tools across multiple package ecosystems on Linux desktops and servers.
- **Developers** — who routinely install CLI tools, runtimes, and libraries from heterogeneous sources and want a single interface.

## Domain boundaries

### In scope

- Searching across all available package backends in parallel.
- Installing, uninstalling, and updating packages via delegated native commands.
- Interactive collision resolution when a package name exists in multiple sources.
- Tracking only packages installed through unipm in a local state file.
- Shell tab-completion for commands, sources, and package names.
- Distrobox container support as a package source.

### Out of scope

- **Project-local dependency management** — no `--dev` / `--save` flags; unipm defaults to user-wide/global installs. This avoids overlapping with language-specific tooling (npm workspaces, pip venvs, etc.).
- **Replacing upstream package managers** — unipm delegates; it does not reimplement registry APIs, dependency resolution, or package building.
- **Running as root** — privilege escalation is scoped to the backend subprocess that requires it (e.g., `sudo apt install`).
- **Source ranking or preference** — unipm has no opinion on which source is "system", "language", or "containerised". It is a pure router.
- **Graphical UI** — unipm is a terminal CLI. The collision prompt uses a TUI (Bubbletea), not a GUI.

## Goals

| # | Goal | Success metric | Target date / phase |
|---|------|----------------|--------------------|
| 1 | Replace context-switching between package managers with a single CLI | Users can complete `search`, `install`, `uninstall`, and `update` without invoking a native package manager directly | Phase 2 (Router & State) |
| 2 | Provide interactive collision resolution | Ambiguous `install` opens a TUI; user resolves in < 3 keystrokes | Phase 2 |
| 3 | Track installed packages for reliable uninstall/update | `unipm uninstall <pkg>` succeeds for any package installed via unipm | Phase 2 |
| 4 | Support the top 7 package ecosystems | `unipm sources` shows apt, npm, pypi, flatpak, brew, pacman/dnf, appimage available | Phase 3 |
| 5 | Distrobox integration | Packages inside distrobox containers are searchable and installable from the host | Phase 4 |
| 6 | Shell tab-completion | `unipm <TAB>` completes commands, sources, and cached package names | Phase 4 |

## Non-goals

| # | Non-goal | Rationale | Revisit? |
|---|----------|-----------|----------|
| 1 | Project-local dependency management (`--dev`, `--save`) | Overlaps with language tooling; adds complexity with venvs and lockfiles | If user demand is overwhelming |
| 2 | Graphical user interface | Terminal-first audience; TUI handles the only interactive need (collision) | If a system tray or GUI wrapper is requested |
| 3 | Remote/daemon mode (e.g., unipm as a service) | Adds attack surface and complexity; local CLI serves the use case | Never — contradicts "local tool" design |
| 4 | Cross-platform (macOS/Windows) in v1 | Backends are Linux-focused; brew is the only cross-platform adapter | Post-v1 if macOS demand exists |

## Goal prioritization

Correctness > Discoverability > Speed > Backend breadth in v1. Every delegated command must succeed or fail with a clear error. Search latency is secondary to result accuracy.

## Stakeholders

| Role | Name | Interest | Influence | Communication | Key concerns |
|------|------|----------|-----------|---------------|--------------|
| Maintainer | Davi | Tool quality, architecture, roadmap | High | Self | Avoiding scope creep; keeping the adapter interface clean |
| Early adopter (sysadmin) | — | Daily package operations, reliability | Medium | GitHub issues | Does it break my system? Is sudo scoped correctly? |
| Early adopter (developer) | — | Fast search, tab-completion, distrobox | Medium | GitHub issues | Can I add my own adapter? Is the config predictable? |

## Assumptions

- The host system runs Linux with at least two of the supported package managers installed.
- The user has `sudo` access for system-level package operations.
- The user is comfortable with a terminal and basic CLI conventions.
- Distrobox users have at least one container already configured.
- Network access is available for registry queries (npm, PyPI, AppImageHub).

## Project phase

Greenfield — no existing system.

## See also

- `users.md` — personas and journeys
- `user_stories.md` — scoped feature requests
