# unipm — Specs Index

> Status: In specification

## Files

| File | Answers |
|------|---------|
| `context.md` | Why are we building this? (problem, goals, stakeholders) |
| `users.md` | Who uses it and how? (personas, journeys) |
| `user_stories.md` | What is in scope? (stories + acceptance criteria) |
| `features/*.feature` | How should each CLI command behave? (Gherkin) |
| `architecture.md` | How is the system structured? (data models, adapter interface, state, config) |
| `adr/` | What did we decide and why? (decision records) |
| `stack.md` | What tools, dependencies, and tests? |
| `ops.md` | How do we distribute and configure it? (installation, config paths, roadmap) |

## Reading order

1. `index.md` (this file)
2. `context.md`
3. `users.md`
4. `user_stories.md`
5. `features/*.feature`
6. `architecture.md`
7. `adr/`
8. `stack.md`
9. `ops.md`

*Why → Who → What → How it behaves → How it's structured → How it's built → How it runs.*

## Rules

- **One concern per file.** If content belongs elsewhere, link out instead of duplicating.
- `.md` files are fill-in templates — replace `[placeholders]` with real content; keep `## Must contain` intact.
- `.feature` files are syntax skeletons — compose freely; no fill-in skeleton is imposed.
- Acceptance criteria use **EARS** (5 patterns). See the skill's `references/ears-quickref.md`; the patterns are also inlined in `user_stories.md`.
- Decisions with trade-offs — including non-obvious algorithms — go in `adr/`.
- Conventions (commits, PRs, naming) live in `CONTRIBUTING.md` and repo config, not in a spec file.

## See also

- `CONTRIBUTING.md` — conventions and how to propose changes
