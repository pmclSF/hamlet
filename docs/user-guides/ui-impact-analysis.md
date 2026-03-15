# UI Impact Analysis Views

## Overview

This document describes the UI views and triage flow for impact analysis in the Terrain extension. The UI is powered by the `ImpactResult` JSON output from the engine — no engine logic is duplicated in the UI layer.

## Changed Area Summary View

The primary impact view shows:

- **Posture badge**: change-risk posture band (well_protected, partially_protected, weakly_protected, high_risk, evidence_limited)
- **File counts**: changed files, source files, test files
- **Impacted units count**: code units affected by the change
- **Protection gap count**: gaps in coverage for changed code
- **Affected owners**: teams with impacted code

## Impacted Tests List

Shows tests relevant to the change, sorted by confidence:

- **Exact tests** (solid border): direct coverage lineage to impacted units
- **Inferred tests** (dashed border): structural or proximity relationship
- **Directly changed** (badge): test file itself was modified

Each test shows `CoversUnits[]` and `Relevance` on expansion.

## Selected Protective Set

The recommended test set view includes:

- **Strategy badge**: exact, near_minimal, or fallback_broad
- **Test list** with selection reasons (expandable per test)
- **Coverage summary**: N covered, M uncovered
- **Gap warning** when uncovered units exist

## Protection Gaps Panel

Groups gaps by severity (high first):

- **Gap type badge**: untested_export, no_coverage, weak_export_coverage, e2e_only_export
- **Path** and **code unit** links
- **Suggested action** for each gap

## Change-Risk Posture Detail

Shows the 4 risk dimensions:

| Dimension | Band | Explanation |
|-----------|------|-------------|
| Protection | ... | ... |
| Exposure | ... | ... |
| Coordination | ... | ... |
| Instability | ... | ... |

Overall band is determined by the worst dimension.

## Triage Flow

The UI guides users through a 4-step triage:

1. **What changed** — changed area summary, file list
2. **What protects it** — impacted tests, selected protective set
3. **What is missing** — protection gaps, coverage diversity issues
4. **What to run first** — protective test set with selection reasoning

## Filtering

- **Exact vs inferred**: toggle to show only exact-confidence impacts
- **Owner/package**: scope to a specific team or directory
- **Severity**: filter gaps by high/medium/low
- **Gap type**: focus on specific gap types

## Confidence Cues

- Impact confidence badges on each code unit and test (exact/inferred/weak)
- Evidence-limited banner when posture is `evidence_limited`
- Limitation callouts from `ImpactResult.Limitations[]`

## Data Source

All views are rendered from the `ImpactResult` JSON structure. The UI never calls the engine directly — it consumes the snapshot or impact result produced by `terrain impact --json` or `terrain pr --json`.
