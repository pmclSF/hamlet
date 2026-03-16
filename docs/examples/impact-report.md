# Example: `terrain impact`

> User question: "What validations matter for this change?"
>
> Terrain impact supports two paths:
> - **Traditional repos:** code change → impacted tests
> - **AI repos:** prompt/dataset change → impacted scenarios

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

---

## Example 2: AI Repo (Prompt Change → Scenarios)

When a prompt or dataset file is changed, Terrain identifies impacted eval scenarios alongside traditional tests.

### Terminal Output

```
Terrain Impact Analysis
============================================================

Summary: 1 file(s) changed, 2 code unit(s) impacted, 1 test(s) relevant, 2 scenario(s) impacted. Posture: partially_protected.

Changed areas:
  src/ai                 src/ai/prompts.ts (modified)

Affected behaviors:
  prompt.ts                            2/2 surfaces changed

Impacted tests:          1
Coverage confidence:     Medium
PR risk:                 PARTIALLY_PROTECTED

Reason categories:
  Directory proximity:     1

Change-Risk Posture: PARTIALLY_PROTECTED
  Some changed code is covered by tests, but gaps remain.
  protection:          partially_protected
  exposure:            partially_protected
  coordination:        well_protected
  instability:         well_protected

Recommended Tests (1)
------------------------------------------------------------
  tests/ai/prompts.test.ts [inferred]
    in same directory tree as changed code

Impacted Scenarios (2)
------------------------------------------------------------
  safety-check (safety) [exact]
    covers 1 changed surface(s)
  accuracy-check (accuracy) [exact]
    covers 2 changed surface(s)

Protection Gaps
------------------------------------------------------------
  [high] Exported function systemPrompt has no observed test coverage.
    Action: Add unit tests for exported function systemPrompt — this is public API surface.

Limitations
------------------------------------------------------------
  * No per-test coverage lineage available; test selection uses structural heuristics.

Next steps:
  terrain impact --show selected   view protective test set with reasoning
  terrain impact --show graph       see dependency graph
  terrain impact --json             machine-readable impact data
```

### JSON Output (`--json`)

```json
{
  "changedAreas": [
    {
      "area": "src/ai",
      "surfaces": [
        {
          "surfaceId": "surface:src/ai/prompts.ts:systemPrompt",
          "name": "systemPrompt",
          "path": "src/ai/prompts.ts",
          "kind": "prompt",
          "changeKind": "modified"
        },
        {
          "surfaceId": "surface:src/ai/prompts.ts:userTemplate",
          "name": "userTemplate",
          "path": "src/ai/prompts.ts",
          "kind": "prompt",
          "changeKind": "modified"
        }
      ]
    }
  ],
  "affectedBehaviors": [
    {
      "behaviorId": "behavior:module:src/ai/prompts.ts",
      "label": "prompt.ts",
      "kind": "module",
      "changedSurfaceCount": 2,
      "totalSurfaceCount": 2
    }
  ],
  "impactedScenarios": [
    {
      "scenarioId": "scenario:eval:safety",
      "name": "safety-check",
      "category": "safety",
      "framework": "promptfoo",
      "relevance": "covers 1 changed surface(s)",
      "impactConfidence": "exact",
      "coversSurfaces": ["surface:src/ai/prompts.ts:systemPrompt"]
    },
    {
      "scenarioId": "scenario:eval:accuracy",
      "name": "accuracy-check",
      "category": "accuracy",
      "framework": "deepeval",
      "relevance": "covers 2 changed surface(s)",
      "impactConfidence": "exact",
      "coversSurfaces": [
        "surface:src/ai/prompts.ts:systemPrompt",
        "surface:src/ai/prompts.ts:userTemplate"
      ]
    }
  ],
  "impactedTestCount": 1,
  "impactedScenarioCount": 2,
  "coverageConfidence": "medium",
  "fallback": {
    "level": "package",
    "reason": "no exact coverage lineage; structural heuristics used",
    "additionalTests": 1
  },
  "postureBand": "partially_protected",
  "summary": "1 file(s) changed, 2 code unit(s) impacted, 1 test(s) relevant, 2 scenario(s) impacted. Posture: partially_protected."
}
```

### Key Differences from Traditional Impact

| Aspect | Traditional | AI Repo |
|--------|------------|---------|
| Changed surfaces | `function`, `method`, `handler`, `route` | `prompt`, `dataset` |
| Impacted validations | Tests (via coverage linkage) | Tests + Scenarios (via CoveredSurfaceIDs) |
| Confidence source | LinkedCodeUnits on test files | Scenario declares covered surfaces explicitly |
| Fallback | directory proximity → full suite | Same ladder applies |

---

## What Engines Contribute

| Section | Engine |
|---------|--------|
| Changed files | `changescope.DetectChanges()` via git diff |
| Changed areas / surfaces | `impact.mapChangedSurfaces()` mapping changed files to CodeSurfaces |
| Impact analysis | `impact.Analyze()` with test linkage and structural heuristics |
| Coverage confidence | Ratio of exact-confidence tests to total impacted tests |
| PR risk / posture | `impact.ScoreRisk()` combining coverage, scope, and ownership dimensions |
| Impacted scenarios | `impact.findImpactedScenarios()` matching changed surfaces to scenario coverage |
| Reason categories | `impact.computeReasonCategories()` classifying test selection reasons |
| Fallback policy | `impact.computeFallbackInfo()` with progressive fallback ladder |
