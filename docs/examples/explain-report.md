# Example: `terrain explain`

> User question: "Why did Terrain select this test?"

## Example 1: Explain a Test (Structured Reason Chain)

```bash
terrain explain test/auth.test.js
```

```
Terrain Explain
============================================================

Target: test/auth.test.js
Framework: jest

Verdict: Selected via exact per-test coverage from src/auth.js:login (confidence: high), plus 1 alternative path(s).

Strongest path (confidence: 95% — high)
------------------------------------------------------------
  src/auth.js:login
    → test/auth.test.js  [exact per-test coverage, confidence: 95%]

Alternative paths (1)
------------------------------------------------------------
  Path 1 (confidence: 95% — high)
  src/auth.js:logout
    → test/auth.test.js  [exact per-test coverage, confidence: 95%]

Confidence: high (95%)
Reason: Direct code dependency

Covers 2 code unit(s):
  src/auth.js:login
  src/auth.js:logout
```

## Example 2: Explain Test Selection Strategy

```bash
terrain explain selection
```

```
Terrain Explain — Test Selection
============================================================

Summary: 1 test(s) selected, strategy: exact, coverage confidence: high.

Strategy: exact
Coverage confidence: high
Tests selected: 1

Reason breakdown:
  Direct code dependency:    1

High confidence (1):
  test/auth.test.js                               95%  via exact per-test coverage from src/auth.js:login

Limitations
------------------------------------------------------------
  * No ownership data available; coordination risk may be underestimated.

Next steps:
  terrain explain <test-path>       explain a specific test
  terrain impact --show selected     view protective test set
  terrain impact --json              machine-readable impact data
```

## Example 3: Explain a Test with Fallback

When no tests directly cover the changed code, the system uses fallback:

```bash
terrain explain selection
```

```
Terrain Explain — Test Selection
============================================================

Summary: 0 test(s) selected, strategy: fallback_broad, coverage confidence: low, 1 protection gap(s).

Strategy: fallback_broad
Coverage confidence: low
Tests selected: 0
Protection gaps: 1

Fallback: all
  no impacted tests identified; full suite fallback

Limitations
------------------------------------------------------------
  * No per-test coverage lineage available; test selection uses structural heuristics.
  * No ownership data available; coordination risk may be underestimated.

Next steps:
  terrain explain <test-path>       explain a specific test
  terrain impact --show selected     view protective test set
  terrain impact --json              machine-readable impact data
```

## Example 4: Explain a Multi-Path Test

When a test covers multiple changed code units, all paths are shown:

```bash
terrain explain test/api.test.js
```

```
Terrain Explain
============================================================

Target: test/api.test.js
Framework: jest

Verdict: Selected via exact per-test coverage from src/api/handler.js:handle (confidence: high), plus 1 alternative path(s).

Strongest path (confidence: 95% — high)
------------------------------------------------------------
  src/api/handler.js:handle
    → test/api.test.js  [exact per-test coverage, confidence: 95%]

Alternative paths (1)
------------------------------------------------------------
  Path 1 (confidence: 95% — high)
  src/api/validator.js:validate
    → test/api.test.js  [exact per-test coverage, confidence: 95%]

Confidence: high (95%)
Reason: Direct code dependency

Covers 2 code unit(s):
  src/api/handler.js:handle
  src/api/validator.js:validate
```

## JSON Output (`--json`)

```bash
terrain explain test/auth.test.js --json
```

```json
{
  "target": {
    "path": "test/auth.test.js",
    "framework": "jest"
  },
  "verdict": "Selected via exact per-test coverage from src/auth.js:login (confidence: high), plus 1 alternative path(s).",
  "strongestPath": {
    "steps": [
      {
        "from": "src/auth.js:login",
        "to": "test/auth.test.js",
        "relationship": "exact per-test coverage",
        "edgeKind": "exact_coverage",
        "edgeConfidence": 0.95
      }
    ],
    "confidence": 0.95,
    "band": "high"
  },
  "alternativePaths": [
    {
      "steps": [
        {
          "from": "src/auth.js:logout",
          "to": "test/auth.test.js",
          "relationship": "exact per-test coverage",
          "edgeKind": "exact_coverage",
          "edgeConfidence": 0.95
        }
      ],
      "confidence": 0.95,
      "band": "high"
    }
  ],
  "confidence": 0.95,
  "confidenceBand": "high",
  "reasonCategory": "directDependency",
  "coversUnits": [
    "src/auth.js:login",
    "src/auth.js:logout"
  ]
}
```

## Auto-Detection and Backward Compatibility

`terrain explain` auto-detects the target type:

1. **`selection`** — explains overall test selection strategy
2. **Test file path** — structured explanation with reason chains (if impact data available), falls back to legacy detail view
3. **Test case ID** — legacy detail view with ID, suite hierarchy, confidence
4. **Code unit** — legacy detail view with covering tests
5. **Owner** — legacy detail view with owned files and signals
6. **Finding** — legacy detail view with type, confidence, explanation

`terrain show test <id>` continues to work unchanged for users who prefer the two-argument form.

## What Engines Contribute

| Section | Engine |
|---------|--------|
| Target metadata | `engine.RunPipeline()` snapshot |
| Reason chains | `impact.BuildImpactGraph()` edges with confidence |
| Confidence scoring | `impact.Confidence` bands (exact/inferred/weak) → numeric scores |
| Fallback detection | `impact.computeFallbackInfo()` from ProtectiveSet strategy |
| Selection strategy | `explain.ExplainSelection()` aggregating per-test explanations |
