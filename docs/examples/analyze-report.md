# Example: `terrain analyze`

> User question: "What is the state of our test system?"

## Terminal Output

```
Terrain — Test Suite Analysis
============================================================

Repository Profile
------------------------------------------------------------
  Test volume:          large
  CI pressure:          high
  Coverage confidence:  medium
  Redundancy level:     medium
  Fanout burden:        high

Validation Inventory
------------------------------------------------------------
  Test files:     1847
  Test cases:     18204
  Code units:     3612
  Code surfaces:  4200
  Frameworks:
    jest                 1145 files [unit]
    vitest                443 files [unit]
    playwright            259 files [e2e]

Coverage Confidence
------------------------------------------------------------
  High:    981 files (54%)
  Medium:  563 files (31%)
  Low:     272 files (15%)

Duplicate Clusters
------------------------------------------------------------
  Clusters:        12
  Redundant tests: 340
  Max similarity:  92%

High-Fanout Nodes
------------------------------------------------------------
  Flagged: 3 (threshold: 10)
    fixtures/authSession.js  (helper, 2400 dependents)
    fixtures/dbSetup.js      (helper, 890 dependents)
    utils/testFactory.js     (helper, 650 dependents)

Skipped Test Burden
------------------------------------------------------------
  Skipped: 187 / 18204 tests (1%)

Weak Coverage Areas
------------------------------------------------------------
  billing-core                               no structural coverage
  checkout/core                              no structural coverage
  legacy/payments                            2 test(s)

CI Optimization Potential
------------------------------------------------------------
  Duplicate tests removable:  340
  Skipped tests reviewable:   187
  High-fanout nodes:          3
  340 duplicate tests could be consolidated; 3 high-fanout nodes could be refactored to reduce blast radius; 187 skipped tests should be reviewed or removed.

Top Insight
------------------------------------------------------------
  fixtures/authSession.js fans out to 2400 transitive dependents — changes here trigger wide test impact. Consider splitting or isolating.

Risk Posture
------------------------------------------------------------
  health:                  STRONG
  coverage_depth:          WEAK
  coverage_diversity:      MODERATE
  structural_risk:         MODERATE
  operational_risk:        STRONG

Signals: 847 total
  42 critical, 156 high, 312 medium, 337 low

Data Completeness
------------------------------------------------------------
  [available ] source
  [available ] coverage
  [missing   ] runtime
  [available ] policy

Limitations
------------------------------------------------------------
  * No runtime data provided; skip/flaky/slow test detection unavailable.

Next steps:
  terrain analyze --verbose           show all findings
  terrain impact                      what tests matter for this change?
  terrain explain <test-path>         why was a test selected?
  terrain insights                    deeper analysis with recommendations
```

## JSON Output (`--json`)

```bash
terrain analyze --json
```

```json
{
  "repository": {
    "name": "my-app",
    "branch": "main",
    "languages": ["javascript", "typescript"],
    "packageManagers": ["npm"],
    "ciSystems": ["github-actions"]
  },
  "dataCompleteness": [
    { "name": "source", "available": true },
    { "name": "coverage", "available": true },
    { "name": "runtime", "available": false },
    { "name": "policy", "available": true }
  ],
  "testsDetected": {
    "testFileCount": 1847,
    "testCaseCount": 18204,
    "codeUnitCount": 3612,
    "codeSurfaceCount": 4200,
    "frameworks": [
      { "name": "jest", "fileCount": 1145, "type": "unit" },
      { "name": "vitest", "fileCount": 443, "type": "unit" },
      { "name": "playwright", "fileCount": 259, "type": "e2e" }
    ]
  },
  "repoProfile": {
    "testVolume": "large",
    "ciPressure": "high",
    "coverageConfidence": "medium",
    "redundancyLevel": "medium",
    "fanoutBurden": "high",
    "manualCoveragePresence": "none",
    "graphDensity": 0.0023
  },
  "coverageConfidence": {
    "highCount": 981,
    "mediumCount": 563,
    "lowCount": 272,
    "totalFiles": 1816
  },
  "duplicateClusters": {
    "clusterCount": 12,
    "redundantTestCount": 340,
    "highestSimilarity": 0.92
  },
  "highFanout": {
    "flaggedCount": 3,
    "threshold": 10,
    "topNodes": [
      { "path": "fixtures/authSession.js", "nodeType": "helper", "transitiveFanout": 2400 },
      { "path": "fixtures/dbSetup.js", "nodeType": "helper", "transitiveFanout": 890 },
      { "path": "utils/testFactory.js", "nodeType": "helper", "transitiveFanout": 650 }
    ]
  },
  "skippedTestBurden": {
    "skippedCount": 187,
    "totalTests": 18204,
    "skipRatio": 0.01
  },
  "weakCoverageAreas": [
    { "path": "billing-core", "testCount": 0, "band": "low" },
    { "path": "checkout/core", "testCount": 0, "band": "low" },
    { "path": "legacy/payments", "testCount": 2, "band": "low" }
  ],
  "ciOptimization": {
    "duplicateTestsRemovable": 340,
    "skippedTestsReviewable": 187,
    "highFanoutNodes": 3,
    "recommendation": "340 duplicate tests could be consolidated; 3 high-fanout nodes could be refactored to reduce blast radius; 187 skipped tests should be reviewed or removed."
  },
  "topInsight": "fixtures/authSession.js fans out to 2400 transitive dependents — changes here trigger wide test impact. Consider splitting or isolating.",
  "riskPosture": [
    { "dimension": "health", "band": "strong" },
    { "dimension": "coverage_depth", "band": "weak" },
    { "dimension": "coverage_diversity", "band": "moderate" },
    { "dimension": "structural_risk", "band": "moderate" },
    { "dimension": "operational_risk", "band": "strong" }
  ],
  "signalSummary": {
    "total": 847,
    "critical": 42,
    "high": 156,
    "medium": 312,
    "low": 337,
    "byCategory": {
      "quality": 312,
      "health": 198,
      "migration": 187,
      "governance": 150
    }
  },
  "limitations": [
    "No runtime data provided; skip/flaky/slow test detection unavailable."
  ]
}
```

## What Engines Contribute

| Section | Engine |
|---------|--------|
| Repository profile | `depgraph.AnalyzeProfile()` |
| Validation inventory | `engine.RunPipeline()` snapshot (tests, scenarios, surfaces) |
| Coverage confidence | `depgraph.AnalyzeCoverage()` |
| Duplicate clusters | `depgraph.DetectDuplicates()` |
| High-fanout nodes | `depgraph.AnalyzeFanout()` |
| Skipped test burden | Signal detection (`skippedTest` signal type) |
| Weak coverage areas | `depgraph.AnalyzeCoverage()` low-band sources |
| CI optimization | Aggregation of duplicates, fanout, and skips |
| Top insight | `analyze.deriveTopInsight()` priority chain |
| Risk posture | `measurement.ComputePosture()` via pipeline measurements |
| Signals | Signal detection across all categories |
| Data completeness | Pipeline data source tracking |
