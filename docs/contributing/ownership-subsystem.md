# Contributing: Ownership Subsystem

This guide covers how to extend, modify, and maintain the ownership subsystem.

## Adding a New Ownership Source

1. Add a new `SourceType` constant in `internal/ownership/model.go`:
   ```go
   SourceGitBlame SourceType = "git_blame"
   ```

2. Add its default confidence in `SourceConfidence()`.

3. Add a loader method on `Resolver` (e.g., `loadGitBlame(root string)`).

4. Add a matcher method that returns `(OwnershipAssignment, bool)`.

5. Insert it at the correct precedence position in `ResolveAssignment()`.

6. Add tests for the new source, including:
   - Happy path
   - Missing/unavailable source
   - Conflict with other sources
   - Edge cases

## Precedence Policy

Sources are evaluated top-to-bottom. The first match wins.

**Do not change precedence order** without careful consideration — existing users rely on explicit config overriding CODEOWNERS. If you need to add a source:

- Higher than CODEOWNERS: place it between explicit config and CODEOWNERS
- Lower than CODEOWNERS: place it between CODEOWNERS and directory fallback

Document any precedence changes in the architecture doc.

## Inheritance Rules

Inheritance flows from coarser to finer entities:

```
file → code unit
file → test case
file → signal (via location)
```

**Key rules:**
- Direct assignments always win over inherited ones
- Inherited assignments preserve the parent's provenance metadata
- Set `Inheritance: InheritanceInherited` on inherited assignments
- Never create circular inheritance (A inherits from B which inherits from A)

## Unowned-Area Handling

- Always use `"unknown"` as the unowned sentinel, never empty string
- Track unowned areas explicitly in summaries and aggregates
- Do not suppress unowned findings — they're often the most important
- When ownership data is entirely absent, report it as a blind spot, not an error

## Ownership-Safe Phrasing

Hamlet's output should be operational, not personal. Follow these patterns:

| Do | Don't |
|----|-------|
| "Risk is concentrated in team-auth's area" | "Team-auth has the worst code" |
| "Instability localized in src/payments/" | "Team-payments is causing failures" |
| "Unowned area has critical findings" | "Nobody cares about this code" |
| "Coordination needed across 4 owner areas" | "Too many teams involved" |

## Testing

- Unit tests: `internal/ownership/*_test.go`
- E2E tests: `internal/testdata/e2e_test.go` (TestE2E_Ownership*)
- All tests use temp directories with real CODEOWNERS files
- No mocks — test real file parsing and resolution

## Privacy Requirements

The benchmark export (`internal/benchmark/export.go`) must never include:
- Owner names or identifiers
- File paths
- Code unit names
- Any data that could identify a specific team or person

Only aggregates: counts, percentages, posture bands, fragmentation indices.

Repos with < 3 owners or < 5 files have certain metrics suppressed to prevent identification.

## File Map

```
internal/ownership/
├── model.go           # Core types, constants, helpers
├── model_test.go      # Model unit tests
├── resolver.go        # Resolution engine
├── resolver_test.go   # Resolver tests
├── codeowners.go      # CODEOWNERS parsing
├── codeowners_test.go # CODEOWNERS tests
├── propagate.go       # Snapshot propagation
├── propagate_test.go  # Propagation tests
├── aggregate.go       # Owner-aware aggregation
└── aggregate_test.go  # Aggregation tests
```
