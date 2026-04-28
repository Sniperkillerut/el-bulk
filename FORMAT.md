# Caveman Encoding Format (FORMAT.md)

This document defines the encoding rules for `SPEC.md` and spec-adjacent documentation.

## Grammar Rules
- Drop articles (a, an, the).
- Drop filler (just, really, basically, simply, actually).
- Drop auxiliary verbs (is, are, was, were, being).
- Drop pleasantries and hedging.
- Use fragments.
- Use short synonyms (fix > implement, big > extensive, run > execute).

## Symbols
- `→` leads to / becomes / on <x>
- `∴` therefore / fix
- `∀` for all / every
- `∃` exists / some
- `!` must / required
- `?` may / optional / unknown
- `⊥` never / forbidden / nil
- `≠` not equal
- `∈` in
- `∉` not in
- `≤` at most
- `≥` at least
- `&` and
- `|` or
- `§` section reference

## Preserve Verbatim
- Code blocks and backticked snippets.
- File paths and URLs.
- Identifiers (function names, variables, env vars).
- Numbers and versions.
- Error strings.
- Structured data (SQL, Regex, JSON, YAML).

## Section Shapes

### Invariants (§V)
`V<n>: <subject> <relation> <condition>`
Example: `V1: ∀ req → auth check before handler`

### Tasks (§T)
Pipe table with columns: `id|status|task|cites`
Statuses: `x` done, `/` wip, `.` todo.
Example: `T3|.|add auth mw|V1,I.api`

### Bugs (§B)
Pipe table with columns: `id|date|cause|fix`
Example: `B1|2026-04-20|token < not ≤|V2`

### Interfaces (§I)
`<kind>: <name> → <shape>`
Example: `api: POST /x → 200 {id:string}`
