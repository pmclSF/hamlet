# Example: `terrain analyze` — Backend API (Go)

> Scenario: A Go backend API with HTTP handlers, route registrations, and go-testing tests. No coverage or runtime artifacts provided.

## Terminal Output

```
Terrain — Test Suite Analysis
============================================================

Repository Profile
------------------------------------------------------------
  Test volume:          small
  CI pressure:          low
  Coverage confidence:  high
  Redundancy level:     low
  Fanout burden:        low

Validation Inventory
------------------------------------------------------------
  Test files:     4
  Test cases:     11
  Code units:     14
  Code surfaces:  10
  Frameworks:
    go-testing              4 files [unit]

Coverage Confidence
------------------------------------------------------------
  High:    3 files (60%)
  Medium:  2 files (40%)
  Low:     0 files (0%)

CI Optimization Potential
------------------------------------------------------------
  Graph has 34 nodes and 57 edges — confidence-based test selection available via `terrain impact`.

Top Insight
------------------------------------------------------------
  No major issues detected. Consider adding coverage or runtime data to unlock deeper analysis.

Risk Posture
------------------------------------------------------------
  health:                  STRONG
  coverage_depth:          STRONG
  coverage_diversity:      STRONG
  structural_risk:         STRONG
  operational_risk:        STRONG

Signals: 0 total

Behavior Redundancy
------------------------------------------------------------
  Redundant tests:  11 across 2 clusters
  [wasteful] 9 tests, 8 shared surfaces (100% confidence) [go-testing]
         Tests in the same framework exercise 8 identical behavior surfaces (100% overlap). Consolidation would reduce CI cost without losing coverage.
  [wasteful] 2 tests, 6 shared surfaces (100% confidence) [go-testing]
         Tests in the same framework exercise 6 identical behavior surfaces (100% overlap). Consolidation would reduce CI cost without losing coverage.

Stability
------------------------------------------------------------
  No runtime data provided. Provide --runtime (JUnit XML or Jest JSON)
  to unlock: flaky test detection, slow test flagging, stability clustering.

Edge Cases
------------------------------------------------------------
  [warning] CI is already fast — optimization may yield minimal benefit.

Policy Recommendations
------------------------------------------------------------
  • CI is already fast. Test selection would yield minimal time savings.

Data Completeness
------------------------------------------------------------
  [available] source
  [missing  ] coverage
  [missing  ] runtime
  [missing  ] policy

Limitations
------------------------------------------------------------
  * No coverage data provided; coverage confidence is structural (import-based) only.
  * No policy file found; governance checks skipped.
  * No runtime data provided; skip/flaky/slow test detection unavailable.

Next steps:
  terrain analyze --verbose           show all findings
  terrain impact                      what tests matter for this change?
  terrain explain <test-path>         why was a test selected?
  terrain insights                    deeper analysis with recommendations
```

## What Each Section Shows

| Section | What It Answers | Key Insight in This Example |
|---------|----------------|----------------------------|
| **Repository Profile** | How big and complex is this test system? | Small Go backend, high structural coverage confidence |
| **Validation Inventory** | What validations exist? | 4 test files, 11 test cases, 10 code surfaces (handlers + functions) |
| **Coverage Confidence** | How well are source files covered? | 60% high, 40% medium — good structural linkage |
| **Behavior Redundancy** | Are tests duplicating effort? | 11 tests in 2 clusters exercising identical behavior surfaces |
| **Stability** | Are tests reliable? | No runtime data — hint to provide `--runtime` |
| **Risk Posture** | Where are the risks? | All strong — healthy backend with good test coverage |

## Backend-Specific Patterns

Terrain detects backend-specific surfaces automatically:

- **HTTP handlers:** Functions matching `*Handler`, `*Middleware`, or accepting `http.ResponseWriter`
- **Route registrations:** `http.HandleFunc("/path")`, `mux.Handle("/path")`, `router.Get("/path")`
- **Exported functions:** Uppercase-first Go functions as code surfaces

These surface kinds enable behavior-aware redundancy detection — tests exercising the same handler through different paths are flagged as overlap candidates.

## Enrichment Opportunities

To unlock deeper analysis for this backend:

```bash
# Add coverage data (Go coverage profile → LCOV)
go test ./... -coverprofile=coverage.out
go tool cover -o coverage.lcov -coverprofile=coverage.out
terrain analyze --coverage coverage.lcov

# Add runtime data (Go test → JUnit XML via gotestsum)
gotestsum --junitfile junit.xml ./...
terrain analyze --runtime junit.xml

# Combined
terrain analyze --coverage coverage.lcov --runtime junit.xml
```
