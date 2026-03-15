# Device Matrix Foundation

> **Status:** Implemented
> **Purpose:** Provide environment and device matrix intelligence — gap detection, concentration analysis, and conservative device recommendations — without overbuilding the full feature.
> **Key decisions:**
> - Matrix analysis operates on the dependency graph, not raw snapshot data
> - Recommendations are conservative: only recommend devices where evidence suggests coverage is relevant
> - A class with zero covered members produces no recommendations (the class may not be relevant)
> - Concentration detection uses a 70% dominance threshold
> - The `terrain matrix` command is deferred; the internal capability is wired into `terrain insights` and `terrain analyze`

**See also:** [18-environment-and-device-model.md](18-environment-and-device-model.md), [02-graph-schema.md](02-graph-schema.md)

## Problem

Terrain models environments and devices as first-class graph nodes (doc 18), but has no analysis layer that answers questions like:

- "Do we have coverage across all supported browsers?"
- "Are tests concentrated on Chrome while Safari and Firefox go untested?"
- "For a given code change, which devices should we also test on?"

Without matrix intelligence, environment data is inert — it exists in the graph but produces no actionable findings.

## Approach

Build a minimal but credible matrix analysis engine (`internal/matrix/`) that operates on the existing environment model and produces:

1. **Class coverage analysis** — per-class breakdown of which members have test coverage
2. **Gap detection** — uncovered members in classes that have some coverage
3. **Concentration detection** — classes where coverage is heavily skewed toward one member
4. **Conservative recommendations** — devices/environments to add, ranked by priority
5. **Change-scoped recommendations** — given impacted tests, recommend representative devices

## Algorithm

### Step 1: Build Member Test Index

Walk all test file nodes in the graph and count `EdgeTargetsEnvironment` outgoing edges to build a reverse index: member ID → test file count.

### Step 2: Analyze Each Class

For each `NodeEnvironmentClass`, walk `EdgeEnvironmentClassContains` edges to find members. Cross-reference with the test index to compute `CoveredMembers` and `CoverageRatio`.

### Step 3: Extract Gaps

A gap is an uncovered member (zero test files targeting it) in a class that has at least one covered member.

### Step 4: Detect Concentration

A class is concentrated when one member holds > 70% of all test file coverage within the class and not all members are covered. This catches "90% of our browser tests target only Chrome" patterns.

### Step 5: Build Recommendations

Priority ordering:
1. **Gaps in partially-covered classes** — these are the most actionable (the class is relevant, but some members are missing)
2. **Underserved members in concentrated classes** — diversify coverage

Classes with zero covered members produce no recommendations — they may not be relevant to the project.

### Change-Scoped Recommendations

`RecommendForTests(g, testPaths)` takes impacted test file paths and:
1. Finds which environments/devices those tests currently target
2. Finds which environment classes contain those targeted members
3. Recommends uncovered members of relevant classes

This enables "if these tests are impacted, also run them on these devices."

## Integration

### `terrain analyze`

The analyze report includes a "Matrix Coverage" section:

```
Matrix Coverage
------------------------------------------------------------
  Classes analyzed:  2
  Test files:        8
  [os] Operating Systems: 2/3 members covered (67%)
  [browser] Browsers: 2/3 members covered (67%)
  Gaps:  2 uncovered members
    - Windows (Operating Systems/os)
    - Firefox 121 (Browsers/browser)
  Recommendations:  2 devices/environments to consider
    1. Windows — No tests target Windows — 2 of 3 os class members have coverage
    2. Firefox 121 — No tests target Firefox 121 — 2 of 3 browser class members have coverage
```

### `terrain insights`

Insights surfaces matrix findings in two categories:

- **Coverage gaps** (coverage_debt): "N environment/device coverage gaps detected"
- **Device concentration** (coverage_debt): "Device concentration: 90% of browser tests target only Chrome"

Both contribute to recommendations and health grade.

### JSON Output

Both `terrain analyze --json` and `terrain insights --json` include the full `MatrixResult`:

```json
{
  "matrixCoverage": {
    "classes": [...],
    "gaps": [...],
    "concentrations": [...],
    "recommendations": [...],
    "testsAnalyzed": 8,
    "classesAnalyzed": 2
  }
}
```

## Implementation

| File | Purpose |
|------|---------|
| `internal/matrix/matrix.go` | Core analysis: `Analyze(g)`, `RecommendForTests(g, paths)`, `FormatSummary(r)` |
| `internal/matrix/matrix_test.go` | 14 tests: class coverage, gaps, concentration, recommendations, empty/nil, determinism, change-scoped, format |
| `internal/insights/insights.go` | `matrixFindings()` emits gap and concentration findings |
| `internal/analyze/analyze.go` | Wires `MatrixCoverage` into analyze report |
| `internal/reporting/analyze_report_v2.go` | Renders "Matrix Coverage" section |
| `internal/reporting/insights_report_v2.go` | Renders matrix summary in insights |
| `cmd/terrain/main.go` | Calls `matrix.Analyze(dg)` in `runInsights()` |

## Future Work: `terrain matrix` Command

A dedicated `terrain matrix` command would provide:

```
terrain matrix                    # Full matrix coverage report
terrain matrix --for <file>       # Change-scoped: recommend devices for impacted tests
terrain matrix --class browser    # Filter to a specific class
```

The internal capability (`matrix.Analyze`, `matrix.RecommendForTests`) is already implemented. The command would be a thin CLI wrapper around these functions.

## Current Scope and Limitations

**What works now:**
- Environment and device nodes exist as first-class graph entities
- Environment classes group related environments/devices
- TestFile → Environment/Device edges connect tests to execution contexts
- Matrix analysis detects gaps, concentrations, and produces recommendations
- Results surface in `terrain insights` and `terrain analyze`

**What is deferred:**
- CI config inference (automatically discovering environments from GitHub Actions matrices, Playwright configs, etc.)
- Runtime history correlation (identifying tests that historically fail on specific devices)
- `terrain matrix` standalone command
- Environment-qualified coverage bands (per-environment coverage scoring)
