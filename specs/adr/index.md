# ADR — Architecture Decision Records

## Must contain

- Each file records one significant decision with: Status, Context, Decision, Alternatives considered, Consequences
- Decisions with trade-offs (incl. non-obvious algorithms) belong here

---

## Directory convention

- Files are numbered sequentially: `NNNN-brief-slug.md`
- `0000-template.md` is the blank template — never delete it.
- Each ADR status: `proposed`, `accepted`, `deprecated`, or `superseded`.
- When a decision is reversed, mark the old ADR `superseded by ADR-NNNN` rather than deleting it.

## ADR index

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [0001](0001-adapter-pattern.md) | Adapter pattern for backend dispatch | Accepted | 2026-07-10 |
| [0002](0002-no-source-priority.md) | No source priority or ranking | Accepted | 2026-07-10 |
| [0003](0003-testing-strategy.md) | Tiered testing strategy with CI offload for containerized backends | Accepted | 2026-07-11 |

## See also

- `architecture.md` — the system this records decisions about
- `stack.md` — technology choices that follow from ADRs
