# ADR-0002: No source priority or ranking

**Status**: Accepted
**Date**: 2026-07-10
**Deciders**: Davi (maintainer)

## Context

When a user runs `unipm install <package>` and the package exists in multiple sources (e.g., `requests` exists in both PyPI and apt), the tool must decide how to present the options. Many meta-package-managers impose a preference — "system packages are preferred over language packages" or "flatpak is preferred over apt."

Adding a ranking system introduces subjective policy into unipm. The maintainer would have to decide that `apt` is "more important" than `npm`, or that `flatpak` is "more sandboxed" and therefore safer. These judgments vary across users, use cases, and even distributions. A ranking that makes sense for a Debian server admin would be wrong for an Arch desktop user.

## Decision

**unipm SHALL NOT rank, categorize, or express preference for any package source.** When a collision occurs, the TUI lists all matching sources alphabetically by source name, with no preselected default. The user must explicitly choose.

The `--source` flag exists precisely for users who know their preference — it bypasses the TUI entirely.

Additionally, unipm SHALL NOT classify sources into categories like "system", "language", or "containerised". All adapters are equal peers. The sources list in the TUI and in `unipm sources` output is sorted alphabetically, not grouped by type.

## Alternatives considered

| Alternative | Pros | Cons | Why rejected |
|-------------|------|------|--------------|
| System-first ranking (apt > flatpak > language PMs) | Quicker for users who always want the system package | Imposes the maintainer's opinion; wrong for many users (e.g., someone who prefers flatpak for sandboxing); creates a "default" that users may not notice | unipm's value is routing, not policy |
| User-configurable priority in config.yaml | Gives users control; satisfies both camps | Adds complexity to config parsing, the TUI (must show "preferred" badges), and documentation; most users would leave defaults anyway | Premature optimization; can revisit if demand is overwhelming |
| Last-used-source memory (auto-select if only one source used before) | Saves keystrokes for repeat installs | Stateful behavior is surprising in a CLI; "which source did I use last time?" is hard to communicate in the TUI; breaks predictability | Violates principle of least surprise |

## Consequences

- **Positive**: unipm remains a pure router — it presents options and lets the user decide. The behavior is predictable and stateless. No config surface is needed for source preferences. The `--source` flag covers the "I know what I want" case.
- **Negative**: Power users who always install from the same source (e.g., "always use brew") must type `--source brew` on every install or use shell aliases. This adds friction for a specific power-user workflow.
- **Mitigations**: The `--source` short flag (`-s`) minimizes keystrokes. Shell aliases (`alias uib='unipm install --source brew'`) are the user's responsibility. If demand is high, a configurable default source can be added later as a non-breaking config extension.
