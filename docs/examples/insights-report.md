# Example: `terrain insights`

> User question: "What should we fix in our test system?"

## Terminal Output

```
Terrain — Test System Health Report
============================================================

Health Grade: C
Headline: 1 critical-severity signals require attention

Repository Profile
------------------------------------------------------------
  Test volume:          large
  CI pressure:          high
  Coverage confidence:  medium
  Redundancy level:     medium
  Fanout burden:        high
  Skip burden:          low
  Flake burden:         medium

Reliability Problems (3)
------------------------------------------------------------
  [CRITICAL] 1 critical-severity signals require attention
         Critical signals indicate issues that are likely to cause test failures or missed regressions.
  [HIGH] 12 flaky/unstable test signals detected
         Flaky tests erode developer trust and waste CI cycles on retries. Stabilize or quarantine these tests.
  [MEDIUM] 187 skipped tests consuming CI resources
         Skipped tests still occupy CI queue slots and mask coverage gaps. Review whether each skip is still justified or should be removed.

Architecture Debt (1)
------------------------------------------------------------
  [HIGH] 3 high-fanout nodes exceed threshold of 10
         Highest: fixtures/authSession.js with 2400 transitive dependents. Changes to high-fanout nodes trigger disproportionate test impact.

Coverage Debt (1)
------------------------------------------------------------
  [HIGH] 272 source files (15%) have low structural coverage
         Files with no test coverage are blind spots for change-scoped test selection. They will not be protected by `terrain impact`.

Optimization Opportunities (1)
------------------------------------------------------------
  [HIGH] 340 redundant tests across 12 clusters
         Largest cluster has 6 tests with 86% similarity. Consolidating duplicates reduces CI runtime and maintenance burden.

Recommended Actions
------------------------------------------------------------
  1. [reliability] 1 critical-severity signals require attention
     why: Critical signals indicate issues that are likely to cause test failures or missed regressions.
     impact: reduced risk of missed regressions
  2. [reliability] 12 flaky/unstable test signals detected
     why: Flaky tests erode developer trust and waste CI cycles on retries. Stabilize or quarantine these tests.
  3. [architecture] Refactor high-fanout fixtures to reduce blast radius
     why: fixtures/authSession.js fans out to 2400 dependents — splitting it isolates test impact.
     impact: narrower test impact per change
  4. [coverage] Add tests for 272 uncovered source files
     why: Coverage gaps mean changes in these files cannot trigger targeted test selection.
     impact: improved change-scoped test selection accuracy
  5. [optimization] Consolidate 12 duplicate test clusters
     why: Removing redundant tests reduces CI runtime and maintenance overhead.
     impact: ~340 fewer tests to maintain

Edge Cases
------------------------------------------------------------
  [caution] High test duplication detected — consider consolidating redundant tests before optimizing.
  [caution] 43% of nodes have excessive fanout — fragile test architecture.

Policy Recommendations
------------------------------------------------------------
  • High test duplication detected. Consider consolidating redundant tests to reduce CI noise.
  • High-fanout fixtures create fragile dependencies. Consider breaking down shared fixtures.

Data Completeness
------------------------------------------------------------
  [available ] source
  [available ] coverage
  [missing   ] runtime
  [available ] policy

Limitations
------------------------------------------------------------
  * No runtime data; flaky/slow test detection relies on structural signals only.

Next steps:
  terrain analyze                     full suite analysis
  terrain impact                      what tests matter for this change?
  terrain explain <test-path>         why was a test selected?
```

## JSON Output (`--json`)

```bash
terrain insights --json
```

```json
{
  "headline": "1 critical-severity signals require attention",
  "healthGrade": "C",
  "findings": [
    {
      "title": "1 critical-severity signals require attention",
      "description": "Critical signals indicate issues that are likely to cause test failures or missed regressions.",
      "category": "reliability",
      "severity": "critical",
      "priority": 1,
      "metric": "1 critical"
    },
    {
      "title": "12 flaky/unstable test signals detected",
      "description": "Flaky tests erode developer trust and waste CI cycles on retries.",
      "category": "reliability",
      "severity": "high",
      "priority": 2,
      "metric": "12 flaky signals"
    },
    {
      "title": "3 high-fanout nodes exceed threshold of 10",
      "description": "Highest: fixtures/authSession.js with 2400 transitive dependents.",
      "category": "architecture_debt",
      "severity": "high",
      "priority": 3,
      "scope": "fixtures/authSession.js",
      "metric": "3 nodes flagged"
    },
    {
      "title": "272 source files (15%) have low structural coverage",
      "description": "Files with no test coverage are blind spots for change-scoped test selection.",
      "category": "coverage_debt",
      "severity": "high",
      "priority": 4,
      "metric": "272/1816 files uncovered"
    },
    {
      "title": "340 redundant tests across 12 clusters",
      "description": "Largest cluster has 6 tests with 86% similarity.",
      "category": "optimization",
      "severity": "high",
      "priority": 5,
      "metric": "340 duplicates"
    },
    {
      "title": "187 skipped tests consuming CI resources",
      "description": "Skipped tests still occupy CI queue slots and mask coverage gaps.",
      "category": "reliability",
      "severity": "medium",
      "priority": 6,
      "metric": "187 skipped"
    }
  ],
  "recommendations": [
    {
      "action": "1 critical-severity signals require attention",
      "rationale": "Critical signals indicate issues that are likely to cause test failures or missed regressions.",
      "category": "reliability",
      "priority": 1,
      "impact": "reduced risk of missed regressions"
    },
    {
      "action": "Refactor high-fanout fixtures to reduce blast radius",
      "rationale": "fixtures/authSession.js fans out to 2400 dependents — splitting it isolates test impact.",
      "category": "architecture_debt",
      "priority": 2,
      "impact": "narrower test impact per change"
    },
    {
      "action": "Add tests for 272 uncovered source files",
      "rationale": "Coverage gaps mean changes in these files cannot trigger targeted test selection.",
      "category": "coverage_debt",
      "priority": 3,
      "impact": "improved change-scoped test selection accuracy"
    },
    {
      "action": "Consolidate 12 duplicate test clusters",
      "rationale": "Removing redundant tests reduces CI runtime and maintenance overhead.",
      "category": "optimization",
      "priority": 4,
      "impact": "~340 fewer tests to maintain"
    }
  ],
  "categorySummary": {
    "reliability": { "count": 3, "criticalCount": 1, "highCount": 1, "topFinding": "1 critical-severity signals require attention" },
    "architecture_debt": { "count": 1, "criticalCount": 0, "highCount": 1, "topFinding": "3 high-fanout nodes exceed threshold of 10" },
    "coverage_debt": { "count": 1, "criticalCount": 0, "highCount": 1, "topFinding": "272 source files (15%) have low structural coverage" },
    "optimization": { "count": 1, "criticalCount": 0, "highCount": 1, "topFinding": "340 redundant tests across 12 clusters" }
  },
  "repoProfile": {
    "testVolume": "large",
    "ciPressure": "high",
    "coverageConfidence": "medium",
    "redundancyLevel": "medium",
    "fanoutBurden": "high",
    "skipBurden": "low",
    "flakeBurden": "medium",
    "graphDensity": 0.0023
  },
  "edgeCases": [
    { "type": "REDUNDANT_TEST_SUITE", "severity": "caution", "description": "High test duplication detected — consider consolidating redundant tests before optimizing." },
    { "type": "HIGH_FANOUT_FIXTURE", "severity": "caution", "description": "43% of nodes have excessive fanout — fragile test architecture." }
  ],
  "policy": {
    "fallbackLevel": "PackageTests",
    "confidenceAdjustment": 0.56,
    "recommendations": [
      "High test duplication detected. Consider consolidating redundant tests to reduce CI noise.",
      "High-fanout fixtures create fragile dependencies. Consider breaking down shared fixtures."
    ]
  },
  "dataCompleteness": [
    { "name": "source", "available": true },
    { "name": "coverage", "available": true },
    { "name": "runtime", "available": false },
    { "name": "policy", "available": true }
  ],
  "limitations": [
    "No runtime data; flaky/slow test detection relies on structural signals only."
  ]
}
```

## Finding Categories

| Category | What it covers | Example findings |
|----------|---------------|------------------|
| **Reliability** | Flaky tests, skipped tests, stability issues, critical signals | "12 flaky/unstable test signals detected" |
| **Architecture Debt** | High-fanout fixtures, fragile dependencies | "3 high-fanout nodes exceed threshold of 10" |
| **Coverage Debt** | Weak structural coverage, e2e-only dependencies | "272 source files have low structural coverage" |
| **Optimization** | Duplicate tests, CI waste | "340 redundant tests across 12 clusters" |

## What Engines Contribute

| Section | Engine |
|---------|--------|
| Findings (duplicates) | `depgraph.DetectDuplicates()` — fingerprinting + similarity clustering |
| Findings (fanout) | `depgraph.AnalyzeFanout()` — BFS transitive reachability |
| Findings (coverage) | `depgraph.AnalyzeCoverage()` — reverse coverage band analysis |
| Findings (skips/flaky) | Signal detection pipeline (`skippedTest`, `flakyTest` signal types) |
| Recommendations | `insights.Build()` — prioritized by severity then category |
| Health grade | `insights.deriveHealthGrade()` — A–D based on finding severity |
| Edge cases | `depgraph.DetectEdgeCases()` — 7 condition types |
| Policy | `depgraph.ApplyEdgeCasePolicy()` — fallback ladder + confidence adjustment |
| Repo profile | `depgraph.AnalyzeProfile()` — volume, pressure, redundancy, fanout |
