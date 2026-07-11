# unipm — User Stories

## Must contain

- Story identifier and title
- "As a ... I want ... so that ..." format
- Acceptance criteria in EARS notation (patterns inlined below)
- Priority (Must / Should / Could / Won't)
- Persona + journey references
- Dependencies on other stories

---

## EARS patterns (for acceptance criteria)

| Behavior | Pattern | Template |
|----------|---------|----------|
| Always true, no condition | Ubiquitous | The system SHALL [response]. |
| Triggered by action/event | Event-driven | WHEN [trigger] THE system SHALL [response]. |
| Error or edge case | Unwanted | IF [condition] THEN THE system SHALL [response]. |
| Depends on state | State-driven | WHILE [state] THE system SHALL [behavior]. |
| Feature/config gated | Optional feature | WHERE [feature is present] THE system SHALL [behavior]. |

---

## Story backlog

| ID | Story | Persona | Priority | Depends on |
|----|-------|---------|----------|------------|
| US-001 | Search across all package sources | Alex, Jordan | Must | None |
| US-002 | Install a package with collision resolution | Alex, Jordan | Must | US-001 |
| US-003 | Uninstall a tracked package | Alex, Jordan | Must | US-002 |
| US-004 | Update tracked packages | Alex | Should | US-002 |
| US-005 | List available package sources | Alex, Jordan | Must | None |
| US-006 | Shell tab-completion | Jordan | Should | US-001 |

---

## US-001: Search across all package sources

**As a** sysadmin or developer,
**I want** to search for a package by name across all available backends in parallel,
**so that** I can find where a package exists without running separate search commands for each package manager.

**Acceptance criteria**:

- The system SHALL accept a search query as a positional argument. (Ubiquitous)
- WHEN the user runs `unipm search <query>` THE system SHALL query every available adapter in parallel via goroutines. (Event-driven)
- The system SHALL display results in a unified table with columns: Source, Name, Version, Description. (Ubiquitous)
- IF a backend query times out (default 10 s) THEN THE system SHALL show partial results from the remaining backends with a warning for the timed-out source. (Unwanted)
- IF no backends are available THEN THE system SHALL display an error message listing the unavailable adapters. (Unwanted)
- WHILE a search is in progress THE system SHALL accept a `--timeout` flag to override the default 10-second timeout per adapter. (State-driven)

**References**:

- Persona: `users.md` → Alex, Jordan
- Journey: `users.md` → Discover and install a package, phase 1
- Feature: `features/search.feature`

---

## US-002: Install a package with collision resolution

**As a** sysadmin or developer,
**I want** to install a package via unipm and be prompted to choose a source when the package exists in multiple backends,
**so that** I never accidentally install from the wrong ecosystem.

**Acceptance criteria**:

- WHEN the user runs `unipm install <package>` and exactly one backend has the package THE system SHALL install it directly without prompting. (Event-driven)
- WHEN the user runs `unipm install <package>` and multiple backends have the package THE system SHALL open an interactive TUI listing all matching sources alphabetically with no preselected default. (Event-driven)
- WHEN the user runs `unipm install <package> --source <source>` THE system SHALL install directly from the named adapter without any prompt. (Event-driven)
- WHEN the user runs `unipm install <package> --source <src1,src2>` THE system SHALL install from every listed source. (Event-driven)
- IF no backend has the package THEN THE system SHALL display an error with a suggestion to try a different query or check `unipm sources`. (Unwanted)
- IF the named source via `--source` is unavailable THEN THE system SHALL display an error listing available sources. (Unwanted)
- The system SHALL record the installed package in `~/.unipm/state.json` with name, source, version, and timestamp. (Ubiquitous)
- The system SHALL delegate the actual installation to the native backend's subprocess (e.g., `sudo apt install`, `npm install -g`). (Ubiquitous)

**References**:

- Persona: `users.md` → Alex, Jordan
- Journey: `users.md` → Discover and install a package, phases 2–4
- Feature: `features/install.feature`

---

## US-003: Uninstall a tracked package

**As a** sysadmin or developer,
**I want** to uninstall a package via unipm and have it automatically route to the correct backend,
**so that** I don't need to remember which package manager I used to install it.

**Acceptance criteria**:

- WHEN the user runs `unipm uninstall <package>` THE system SHALL look up the package in `~/.unipm/state.json`. (Event-driven)
- WHEN the package is found in the state file THE system SHALL delegate the removal to the recorded source's backend (e.g., `sudo apt remove`, `pip3 uninstall`). (Event-driven)
- The system SHALL remove the package record from `state.json` after successful removal. (Ubiquitous)
- IF the package is not found in the state file THEN THE system SHALL display an error: "<package> was not installed via unipm". (Unwanted)
- IF the backend removal command fails THEN THE system SHALL display the native error and offer to clean the state record. (Unwanted)

**References**:

- Persona: `users.md` → Jordan
- Journey: `users.md` → Remove an installed package, phases 1–3
- Feature: `features/uninstall.feature`

---

## US-004: Update tracked packages

**As a** sysadmin,
**I want** to update packages I installed via unipm, either individually or all at once,
**so that** my system stays current without checking each backend manually.

**Acceptance criteria**:

- WHEN the user runs `unipm update` without arguments THE system SHALL iterate all packages in `state.json` and delegate updates to each package's recorded backend. (Event-driven)
- WHEN the user runs `unipm update <package>` THE system SHALL update only the named package if it exists in the state file. (Event-driven)
- The system SHALL refresh the `version` field in `state.json` after each successful update. (Ubiquitous)
- IF a package in the state file fails to update THEN THE system SHALL report the failure per-package and continue with the remaining packages. (Unwanted)
- IF the named package is not found in the state file THEN THE system SHALL display an error: "<package> was not installed via unipm". (Unwanted)

**References**:

- Persona: `users.md` → Alex
- Feature: `features/update.feature`

---

## US-005: List available package sources

**As a** sysadmin or developer,
**I want** to see which package sources are detected and available on my system,
**so that** I know which backends unipm can use before I search or install.

**Acceptance criteria**:

- WHEN the user runs `unipm sources` THE system SHALL display every compiled-in adapter with its status. (Event-driven)
- The system SHALL mark each adapter as "✓ available" or "✗ not found on $PATH". (Ubiquitous)
- The system SHALL list distrobox adapters per configured container and mark them based on container existence. (Ubiquitous)
- IF no adapters are available THEN THE system SHALL display a message suggesting the user install at least one supported package manager. (Unwanted)

**References**:

- Persona: `users.md` → Alex, Jordan
- Feature: `features/sources.feature`

---

## US-006: Shell tab-completion

**As a** developer,
**I want** tab-completion for commands, source names, and package names,
**so that** I can use unipm as fast as I type.

**Acceptance criteria**:

- The system SHALL generate completion scripts for bash, zsh, and fish via `unipm completion <shell>`. (Ubiquitous)
- WHEN the user presses `<TAB>` after `unipm install --source ` THE system SHALL complete with the list of available adapter names. (Event-driven)
- WHEN the user presses `<TAB>` after `unipm install ` THE system SHALL complete with cached package names from previous searches. (Event-driven)
- WHERE a local package-name cache exists at `~/.unipm/cache.json` THE system SHALL serve completions from it with a configurable TTL (default 24 h). (Optional feature)
- IF the user's input is fewer than 3 characters THEN THE system SHALL not attempt network-backed completions. (Unwanted)

**References**:

- Persona: `users.md` → Jordan
- Feature: `features/completion.feature`

## See also

- `users.md` — personas and journeys
- `features/*.feature` — Gherkin scenarios per story
- `architecture.md` — implementation of stories
