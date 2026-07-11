# unipm — Users

## Must contain

- Persona: name + tagline, background, goals, pain points, technical proficiency, scenarios
- Journey: name, persona, trigger, goal, phased steps (action / system response / touchpoint / sentiment / pain)

---

## Personas

### Alex — System Administrator

**Tagline**: "I maintain 20 workstations and I'm tired of remembering which package manager has what."

**Background**: Alex manages a fleet of Linux desktops for a small engineering team. They install tools daily — monitoring agents, CLI utilities, development runtimes — and the context-switching between `apt`, `snap`, `flatpak`, and `brew` is a constant source of friction. Alex writes shell scripts for automation and values predictable, scriptable interfaces.

**Goals**:

- Install any package with a single command, regardless of its source ecosystem.
- Uninstall packages without remembering which manager was used originally.
- Script package operations with predictable, non-interactive flags (`--source`).

**Pain points**:

- "I ran `apt remove` on a package I installed with `flatpak` — nothing happened and I wasted 5 minutes."
- "Searching for a tool means running `apt search`, then `flatpak search`, then `brew search` — and comparing results by hand."

**Technical proficiency**: High — comfortable with terminal, shell scripting, `sudo`, and Linux package management internals.

**Scenarios**:

1. Alex needs to install `ripgrep` across 20 machines. They want one command that works regardless of whether the source is `apt`, `brew`, or `cargo`.
2. Alex discovers `bat` (a `cat` replacement) and wants to find it quickly without guessing whether it's an `apt` package, a `brew` formula, or a Rust crate.

### Jordan — Developer

**Tagline**: "I live in the terminal. Don't make me leave it to search for a package."

**Background**: Jordan is a full-stack developer who runs Arch Linux with distrobox containers for isolated development environments. They install tools from `pacman`, `yay` (AUR), `npm`, `pip`, and occasionally `flatpak`. Jordan values speed, tab-completion, and not breaking their carefully configured system.

**Goals**:

- Search all registries in parallel from one command.
- Install packages into distrobox containers from the host terminal.
- Never manually track "where did I install that tool?" again.

**Pain points**:

- "I have an AUR helper, npm, pip, and flatpak — that's four different search syntaxes."
- "I installed `httpie` via pip six months ago. Was it pip3? pip? --user? I can't remember."
- "Tab-completing package names across different managers is impossible."

**Technical proficiency**: High — daily terminal user, comfortable with Arch, AUR, containers, and scripting.

**Scenarios**:

1. Jordan wants to try a new JSON processor. They search once, see results from `pacman`, `npm`, and `brew`, and pick the fastest install path.
2. Jordan has an `arch-dev` distrobox container and wants to install `neovim` inside it without entering the container manually.

## Journeys

### Discover and install a package

**Persona**: Alex (Sysadmin) and Jordan (Developer)
**Trigger**: The user wants to install a tool they know by name (e.g., `ripgrep`) or by function (e.g., "JSON processor").
**Goal**: Find the package, understand which sources have it, and install it — with minimal keystrokes.

| Phase | User action | System response | Touchpoint | Sentiment | Pain point / Opportunity |
|-------|-------------|-----------------|------------|-----------|--------------------------|
| 1. Discovery | User runs `unipm search ripgrep` | unipm fans out to all available backends in parallel, merges results, and displays a colourised table: `Source | Name | Version | Description` | Terminal | 😊 | If a backend times out, show partial results with a warning rather than failing entirely |
| 2. Decision | User reads the table and decides to install | If exactly one match exists, unipm proceeds silently. If multiple matches exist, unipm opens a TUI list for selection (alphabetical, no preselected default) | TUI (Bubbletea) | 😐 | The TUI must be navigable with j/k/arrows and selectable with Enter — no mouse required |
| 3. Installation | User selects a source or confirms the single match | unipm delegates to the native backend (e.g., `sudo apt install ripgrep`), streams output, and records the result in `~/.unipm/state.json` | Terminal | 😊 | Scoped `sudo` — only the apt/pacman subprocess runs elevated, not unipm itself |
| 4. Confirmation | User sees the installation output | unipm prints a summary line: "✓ ripgrep 14.1.0 installed from apt" | Terminal | 😊 | If installation fails, show the native error and suggest `unipm sources` to verify backend availability |

### Remove an installed package

**Persona**: Jordan (Developer)
**Trigger**: User no longer needs a package installed via unipm.
**Goal**: Remove the package completely, regardless of which backend originally installed it.

| Phase | User action | System response | Touchpoint | Sentiment | Pain point / Opportunity |
|-------|-------------|-----------------|------------|-----------|--------------------------|
| 1. Recall | User runs `unipm uninstall httpie` | unipm looks up `httpie` in `~/.unipm/state.json`, finds the source (`pypi`), and delegates `pip3 uninstall httpie` | Terminal | 😊 | If the package is not in the state file, error with a clear message: "httpie was not installed via unipm" |
| 2. Removal | — (no further input needed) | Backend removes the package; unipm removes the record from state.json | Terminal | 😊 | If the backend removal fails (e.g., package already manually removed), unipm should still offer to clean the state record |
| 3. Confirmation | User sees removal output | "✓ httpie removed from pypi" | Terminal | 😊 | — |

## See also

- `context.md` — stakeholders (superset of personas)
- `user_stories.md` — stories carved from journey phases
- `features/*.feature` — behavior per touchpoint
