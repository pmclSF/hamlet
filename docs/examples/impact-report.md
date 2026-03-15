# Example: `terrain impact`

> User question: "What tests matter for this change?"

## Terminal Output

```
Terrain Impact Analysis
============================================================

Summary: 2 file(s) changed, 4 code unit(s) impacted, 2 test(s) relevant. Posture: partially_protected.

Changed areas:
  auth                   src/auth/token_validation.ts (modified)
  auth                   src/auth/session_manager.ts (modified)

Impacted tests:          2
Coverage confidence:     Medium
PR risk:                 PARTIALLY_PROTECTED

Reason categories:
  Direct code dependency:  1
  Fixture dependency:      1

Change-Risk Posture: PARTIALLY_PROTECTED
  Some changed code is covered by tests, but gaps remain.
  test_coverage:         medium
  change_scope:          moderate
  ownership_spread:      single_owner

Recommended Tests (2)
------------------------------------------------------------
  tests/auth/login.spec.ts [high]
    imports src/auth/token_validation.ts
  tests/integration/auth-flow.spec.ts [medium]
    uses fixture that depends on src/auth/session_manager.ts

Protection Gaps
------------------------------------------------------------
  [medium] src/auth/session_manager.ts:refreshSession has no direct test coverage
    Action: Add unit test for refreshSession

Limitations
------------------------------------------------------------
  * No ownership data available; coordination risk may be underestimated.

Next steps:
  terrain impact --show selected   view protective test set with reasoning
  terrain impact --show graph       see dependency graph
  terrain impact --json             machine-readable impact data
```

## JSON Output (`--json`)

```json
{
  "changedAreas": [
    {
      "area": "auth",
      "surfaces": [
        {
          "surfaceId": "surface:src/auth/token_validation.ts:validateToken",
          "name": "validateToken",
          "path": "src/auth/token_validation.ts",
          "kind": "function",
          "changeKind": "modified"
        },
        {
          "surfaceId": "surface:src/auth/session_manager.ts:refreshSession",
          "name": "refreshSession",
          "path": "src/auth/session_manager.ts",
          "kind": "function",
          "changeKind": "modified"
        }
      ]
    }
  ],
  "impactedUnitCount": 4,
  "impactedTestCount": 2,
  "protectionGapCount": 1,
  "selectedTestCount": 2,
  "coverageConfidence": "medium",
  "reasonCategories": {
    "directDependency": 1,
    "fixtureDependency": 1,
    "directlyChanged": 0,
    "directoryProximity": 0
  },
  "fallback": {
    "level": "none",
    "additionalTests": 0
  },
  "postureBand": "partially_protected",
  "summary": "2 file(s) changed, 4 code unit(s) impacted, 2 test(s) relevant. Posture: partially_protected.",
  "limitations": [
    "No ownership data available; coordination risk may be underestimated."
  ]
}
```

## What Engines Contribute

| Section | Engine |
|---------|--------|
| Changed files | `changescope.DetectChanges()` via git diff |
| Changed areas / surfaces | `impact.mapChangedSurfaces()` mapping changed files to CodeSurfaces |
| Impact analysis | `impact.Analyze()` with test linkage and structural heuristics |
| Coverage confidence | Ratio of exact-confidence tests to total impacted tests |
| PR risk / posture | `impact.ScoreRisk()` combining coverage, scope, and ownership dimensions |
| Reason categories | `impact.computeReasonCategories()` classifying test selection reasons |
| Fallback policy | `impact.computeFallbackInfo()` with progressive fallback ladder |
