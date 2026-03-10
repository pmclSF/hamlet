# Ownership Architecture

## Overview

Ownership is a **routing layer** in Hamlet, not a blame layer. It connects technical findings (risk, health, quality, migration) to the people and teams who can act on them.

The ownership subsystem makes Hamlet's output more actionable by answering: *who should look at this?*

## Core Principles

1. **Routing, not blame.** Ownership exists to route findings to teams, not to create scorecards or rankings.
2. **Explicit and inspectable.** Every ownership assignment carries provenance (source, confidence, matched rule).
3. **Unknown is first-class.** Unowned areas are visible, not silently hidden.
4. **Multi-owner support.** Shared ownership is a real pattern, not an edge case.
5. **Inherited vs direct.** A code unit inheriting ownership from its file is weaker evidence than a CODEOWNERS rule.

## Data Model

### Core Types (`internal/ownership/model.go`)

| Type | Purpose |
|------|---------|
| `Owner` | Canonical owner identity (ID + display name) |
| `OwnershipAssignment` | Full assignment with owners, source, confidence, inheritance |
| `SourceType` | Where the assignment came from (CODEOWNERS, explicit config, etc.) |
| `Confidence` | How trustworthy the assignment is (high, medium, low, none) |
| `InheritanceKind` | Whether ownership is direct or inherited from a parent |
| `EntityType` | What kind of entity is owned (file, module, test case, code unit, etc.) |
| `OwnerAggregate` | Per-owner summary statistics |
| `OwnershipSummary` | Snapshot-level ownership overview |
| `Diagnostic` | Warning or issue from ownership resolution |

### Source Precedence

Resolution evaluates sources in this order (highest to lowest):

1. **Explicit config** (`.hamlet/ownership.yaml` rules) — `ConfidenceHigh`
2. **CODEOWNERS** (standard GitHub locations) — `ConfidenceHigh`
3. **Path mappings** (`.hamlet/ownership.yaml` path_mappings) — `ConfidenceMedium`
4. **Git history fallback** (`.hamlet/ownership.yaml` git_history, or auto when CODEOWNERS is absent) — `ConfidenceLow`
5. **Directory fallback** (top-level directory name) — `ConfidenceLow`
6. **Unknown** — `ConfidenceNone`

The first source that matches wins. Within CODEOWNERS, the last matching rule wins (GitHub convention).

### Inheritance Rules

| Parent Entity | Child Entity | Rule |
|--------------|-------------|------|
| File | CodeUnit | Code unit inherits from file unless directly assigned |
| Test file | TestCase | Test case inherits from its file |
| File | Signal | Signal inherits from its location file |
| File/Module | CoverageInsight | Insight inherits from its file |

Direct assignments always override inherited ones.

## Architecture

```
┌─────────────────────────────┐
│  Ownership Sources          │
│  ┌─────────────────────┐    │
│  │ .hamlet/ownership.yaml│   │
│  │ CODEOWNERS            │   │
│  │ Path mappings         │   │
│  │ Git history fallback  │   │
│  └─────────────────────┘    │
└──────────┬──────────────────┘
           │
           ▼
┌─────────────────────────────┐
│  Resolver                   │
│  - NewResolver(root)        │
│  - Resolve(path) → string   │
│  - ResolveAssignment(path)  │
│    → OwnershipAssignment    │
└──────────┬──────────────────┘
           │
           ▼
┌─────────────────────────────┐
│  Propagate()                │
│  - Phase 1: Test files      │
│  - Phase 2: Code units      │
│  - Phase 3: Signals         │
│  - Phase 4: Ownership map   │
│  - Phase 5: Summary         │
└──────────┬──────────────────┘
           │
           ▼
┌─────────────────────────────────────────────────┐
│  Downstream Consumers                           │
│  ┌──────────┐ ┌──────────┐ ┌─────────────────┐ │
│  │ Heatmap  │ │ Summary  │ │ Benchmark Export│ │
│  │ (owner   │ │ (focus,  │ │ (privacy-safe   │ │
│  │ hotspots)│ │ blinds)  │ │ aggregates)     │ │
│  └──────────┘ └──────────┘ └─────────────────┘ │
│  ┌──────────┐ ┌──────────┐ ┌─────────────────┐ │
│  │ Compare  │ │ Migration│ │ Review/Analyze  │ │
│  │ (owner   │ │ (coord   │ │ (by-owner       │ │
│  │ trends)  │ │ risk)    │ │ grouping)       │ │
│  └──────────┘ └──────────┘ └─────────────────┘ │
└─────────────────────────────────────────────────┘
```

## Pipeline Integration

Ownership propagation runs as **Step 4** in the engine pipeline (`internal/engine/pipeline.go`), after signal detection and before risk scoring:

```go
resolver := ownership.NewResolver(root)
ownership.Propagate(resolver, snapshot)
```

This populates:
- `TestFile.Owner` on all test files
- `CodeUnit.Owner` on all code units (inherited from file)
- `Signal.Owner` on all signals (inherited from location file)
- `TestSuiteSnapshot.Ownership` map for downstream consumers

## Owner-Aware Aggregation

The `ownership/aggregate.go` module provides:

| Function | Purpose |
|----------|---------|
| `BuildHealthSummaries()` | Flaky/slow/skipped tests by owner |
| `BuildQualitySummaries()` | Untested exports, weak assertions by owner |
| `BuildMigrationSummaries()` | Migration blockers by owner |
| `ComputeMigrationCoordinationRisk()` | How many owners are affected by migration |
| `CompareOwnerSignals()` | Signal trend by owner across snapshots |
| `BuildFocusItems()` | Ownership-aware focus recommendations |
| `BuildBenchmarkAggregate()` | Privacy-safe ownership stats |

## Privacy Boundary

The benchmark export never includes:
- Raw owner names
- Raw file paths
- Raw code unit names
- Small-cell identifying detail

It only includes:
- Owner count
- Coverage posture band
- Top-owner risk share percentage
- Unowned critical code percentage
- Fragmentation index

Repos with fewer than 3 owners or 5 files have fragmentation and risk-share suppressed.

## CODEOWNERS Parsing

Supported patterns:
- `*` — match everything
- `*.js` — wildcard extension
- `/src/auth/` — root-anchored directory prefix
- `src/auth/` — directory prefix
- `docs/*` — single-level wildcard
- `**/test/` — double-star directory match

Unsupported patterns generate diagnostics (not silent failures):
- `[abc]` character classes
- `{a,b}` brace expansion
- `?` single-character wildcard

## Configuration

### `.hamlet/ownership.yaml`

```yaml
ownership:
  rules:
    - path: "packages/payments/"
      owner: "team-payments"
    - path: "packages/auth/"
      owner: "team-auth"
  path_mappings:
    - prefix: "lib/shared/"
      owners: ["team-platform", "team-infra"]
  git_history:
    enabled: true
    max_commits: 1000
    auto_when_no_codeowners: true
```

Rules use longest-prefix matching. Path mappings support multiple owners.
`git_history` uses recent commit author identities as a low-confidence fallback.
When no CODEOWNERS file is present, this fallback is enabled by default and can be
disabled with `auto_when_no_codeowners: false`.

## File Inventory

| File | Purpose |
|------|---------|
| `internal/ownership/model.go` | Core types and constants |
| `internal/ownership/resolver.go` | Resolution engine with precedence |
| `internal/ownership/codeowners.go` | CODEOWNERS parsing and matching |
| `internal/ownership/propagate.go` | Snapshot-wide ownership propagation |
| `internal/ownership/aggregate.go` | Owner-aware health/quality/migration aggregation |
| `internal/ownership/*_test.go` | Tests for all modules |
