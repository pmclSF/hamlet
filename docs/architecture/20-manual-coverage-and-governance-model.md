# Manual Coverage and Governance Model

> **Status:** Implemented
> **Purpose:** Define how Terrain models manual test coverage, QA processes, and governance overlays that exist outside CI.
> **Key decisions:**
> - Manual coverage is modeled as graph nodes (ManualCoverageArtifact), not metadata annotations
> - Manual coverage supplements but never replaces automated coverage
> - Governance evaluation is local-first — policy rules are evaluated against the local graph, not a remote service
> - Manual coverage carries a confidence adjustment (0.9x) relative to automated coverage
> - ManualCoverageArtifact implements `ValidationTarget` — the shared interface that unifies tests, scenarios, and manual coverage artifacts (see `internal/models/validation_target.go`). `IsExecutable()` always returns `false` for manual coverage.

**See also:** [02-graph-schema.md](02-graph-schema.md), [12-risk-and-coverage-taxonomy.md](12-risk-and-coverage-taxonomy.md), [15-edge-case-handling.md](15-edge-case-handling.md)

## Problem

Many teams maintain validation processes that happen entirely outside automated CI. TestRail regression suites, QA checklists, release sign-off procedures, manual exploratory testing sessions — these represent real coverage over real behavior surfaces, but they are invisible to any tool that only analyzes code and CI artifacts.

This creates a distorted picture. A team that runs a rigorous weekly regression suite in TestRail and employs dedicated QA engineers for exploratory testing may appear to have poor coverage if Terrain only considers automated tests. The risk scores become misleading, and the recommendations become noise.

Terrain must model manual coverage as a first-class concept so that risk assessment reflects the full validation landscape.

## Current Implementation

### ManualCoverageArtifact Nodes

Manual coverage entries are represented as `ManualCoverageArtifact` nodes in the dependency graph. Each node carries:

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Human-readable name of the manual coverage activity |
| `area` | string | Code area or behavior surface this coverage applies to (e.g., `auth/login`, `checkout/payment`) |
| `source` | enum | Origin system: `testrail`, `jira`, `qase`, `checklist`, `exploratory`, `manual` |
| `owner` | string | Team or individual responsible for executing this coverage |
| `criticality` | enum | `high`, `medium`, `low` — how critical this manual coverage is to release confidence |
| `lastExecuted` | timestamp | When this manual coverage was last executed (optional, used for staleness detection) |
| `frequency` | string | Expected execution cadence: `per-release`, `weekly`, `monthly`, `ad-hoc` |

### Configuration via terrain.yaml

Manual coverage is declared in `terrain.yaml` under the `manual_coverage` key. The file is loaded by `policy.LoadTerrainConfig()` during the pipeline's preparation phase. Entries are converted to `ManualCoverageArtifact` values via `TerrainConfig.ToManualCoverageArtifacts()` and appended to the snapshot before signal detection runs.

```yaml
manual_coverage:
  - name: Login regression suite
    area: auth/login
    source: testrail
    owner: qa-team
    criticality: high
    frequency: per-release

  - name: Payment flow exploratory testing
    area: checkout/payment
    source: exploratory
    owner: senior-qa
    criticality: high
    frequency: weekly

  - name: Admin panel smoke test
    area: admin/*
    source: checklist
    owner: dev-team
    criticality: medium
    frequency: per-release
```

Each entry requires `name` and `area`. Missing `source` defaults to `"manual"`, missing `criticality` defaults to `"medium"`. The artifact ID is deterministic: `manual:<source>:<sha256(lowercase(name))[:16]>`.

### Graph Integration

ManualCoverageArtifact nodes are built into the dependency graph via `buildManualCoverage()` in `internal/depgraph/build.go`. Each artifact becomes a `NodeManualCoverage` node with metadata (source, criticality, frequency, lastExecuted, area). Covered surfaces are connected via `EdgeManualCovers` edges (confidence 0.7, evidence `manual`). If an owner is specified, a `NodeOwner` node is created with an `EdgeOwns` edge.

Manual coverage nodes are queryable via `Graph.ValidationTargets()` and `Graph.ValidationsForSurface()` — the same queries used for tests and scenarios. The `IsValidationNode()` predicate returns true for `NodeManualCoverage`.

#### Attachment: Explicit vs. Area-Resolved

Manual coverage can attach to graph entities in two ways:

1. **Explicit `surfaces` list** — when `CoveredSurfaceIDs` is provided, edges are created directly to the listed node IDs with confidence 0.7.

2. **Area-based resolution** — when no explicit surfaces are listed, the `area` field is matched against packages, services, behavior surfaces, and code surfaces in the graph. The `resolveAreaToGraphNodes()` function checks each attachable node's `Package`, `Path`, and `Name` fields against the area string. This means a single area like `"billing"` can automatically attach to all billing-related packages, services, and behavior surfaces.

Edge confidence for area-resolved matches depends on specificity:

- Exact match on package or name — confidence 0.7
- Glob/prefix match (e.g., `auth/*`) — confidence 0.5
- Feature-area prefix match — confidence 0.5

Explicit surfaces always take priority: when `CoveredSurfaceIDs` is non-empty, area resolution is skipped entirely.

### Command Integration

Manual coverage is surfaced in three Terrain commands:

**`terrain analyze`** — The analyze report includes a "Manual Coverage Overlay" section showing artifact count, sources, criticality breakdown, covered areas, and stale artifact count. The repository profile shows `manualCoveragePresence` (none/partial/significant).

**`terrain insights`** — The insights health report includes manual coverage presence in the repository profile. When half or more manual coverage artifacts lack a `lastExecuted` date, a staleness finding is raised under the coverage debt category.

**`terrain impact`** — When protection gaps overlap with manually covered areas, impact analysis annotates the result with policy notes indicating that manual coverage exists for the area. These notes specify the artifact name, source, and criticality, and remind users to verify manually.

In all three commands, manual coverage is clearly labeled as non-executable. It never inflates automated test counts or executable coverage metrics.

## How Manual Coverage Affects Risk

### Coverage Gap Reduction

When Terrain identifies a behavior surface with no automated test coverage, it checks for manual coverage nodes that cover the same area. If manual coverage exists, the gap severity is reduced proportionally:

- **High-criticality manual coverage** reduces gap severity by two levels (critical → low)
- **Medium-criticality manual coverage** reduces gap severity by one level (critical → medium)
- **Low-criticality manual coverage** flags the gap as acknowledged but does not reduce severity

### Confidence Adjustment

Manual coverage inherently carries lower confidence than automated coverage. Automated tests are deterministic, repeatable, and verifiable. Manual processes depend on human consistency and may drift from their documented scope.

Terrain applies a **0.9x confidence multiplier** to all manual coverage contributions. This means a behavior surface covered only by manual testing will never reach the same confidence level as one covered by automated tests, even with high-criticality manual coverage.

This multiplier compounds with edge confidence: a directory-match manual coverage entry (0.7 edge confidence) with the manual adjustment (0.9x) produces an effective confidence of 0.63.

### Staleness Detection

If a manual coverage entry includes `lastExecuted` and `frequency`, Terrain detects stale manual coverage — entries that have not been executed within their expected cadence. Stale manual coverage receives an additional confidence penalty:

- One missed cycle: 0.8x additional penalty
- Two or more missed cycles: 0.5x additional penalty
- Three or more missed cycles: manual coverage is flagged as likely abandoned

## Governance Model

### Policy Rules

Governance is defined as policy rules in `terrain.yaml` under the `governance` key:

```yaml
governance:
  policies:
    - name: "Critical paths require automated coverage"
      target: "criticality:high"
      require:
        automatedCoverage: true
        minimumConfidence: 0.7

    - name: "All areas require some coverage"
      target: "*"
      require:
        coverageType: [automated, manual]
        minimumConfidence: 0.5
```

### Policy Evaluation

`terrain policy check` evaluates the combined automated and manual coverage against policy rules. Each rule specifies a target (which behavior surfaces it applies to) and requirements (what coverage must exist).

Policy evaluation produces:

- **Pass** — all requirements met with sufficient confidence
- **Warn** — requirements met but with marginal confidence or stale manual coverage
- **Fail** — requirements not met

### Sign-off Workflows

Governance policies can require explicit sign-off for release. Sign-off rules reference manual coverage owners and require that manual coverage has been executed within the current release cycle. This is evaluated locally against the graph — Terrain does not manage sign-off state, it evaluates whether the conditions for sign-off are met.

## Future: Test Management API Integration

The current implementation requires manual declaration of coverage in `terrain.yaml`. Future phases will integrate with test management platforms to ingest manual coverage data automatically:

- **TestRail** — import test runs and case coverage mappings
- **Xray (Jira)** — import test execution results and requirement links
- **Qase** — import test run results and suite structures

API integration will replace static `terrain.yaml` entries with live data, improving staleness detection and reducing configuration drift. The graph model remains the same — ManualCoverageArtifact nodes are created from API data instead of YAML configuration.

## Key Decisions

1. **Graph nodes, not metadata.** Manual coverage is modeled as first-class nodes in the dependency graph, not annotations on existing nodes. This allows manual coverage to participate in graph traversal, impact analysis, and confidence propagation using the same algorithms as automated coverage.

2. **Supplements, never replaces.** Manual coverage reduces gap severity and contributes to coverage confidence, but a governance policy that requires `automatedCoverage: true` cannot be satisfied by manual coverage alone. This reflects the fundamental reliability difference between automated and manual validation.

3. **Local-first governance.** Policy evaluation runs against the local graph using `terrain policy check`. There is no remote governance service, no approval database, no external state. This keeps governance deterministic and auditable.
