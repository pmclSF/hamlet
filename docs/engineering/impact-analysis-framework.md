# Impact Analysis Framework

## Overview

Impact analysis answers: "If this code changes, which tests matter, what protection exists, and where are the gaps?"

It sits above the existing subsystems:
- CodeUnit inventory
- TestFile/framework detection
- Coverage lineage (LinkedCodeUnits)
- Ownership resolution
- Posture/measurement systems

## Core Concepts

| Concept | Description |
|---------|-------------|
| ChangeScope | What changed: files, code units, tests |
| ImpactedCodeUnit | A code unit affected by the change |
| ImpactedTest | A test relevant to the change |
| ProtectionGap | Where changed code lacks adequate coverage |
| ChangeRiskPosture | Overall risk assessment scoped to the change |
| SelectedTests | Recommended protective test set |

## How it differs from other subsystems

| Subsystem | Scope | Purpose |
|-----------|-------|---------|
| Repo-wide posture | Entire repo | Overall health assessment |
| Impact analysis | Specific change | Change-scoped protection assessment |
| PR reporting | Specific PR | Renders impact results for review |
| Portfolio intelligence | Cross-repo | Aggregate organizational view |

## Change-scope inputs

Changes can come from:
- `git diff` against a base ref
- Explicit file paths
- CI-provided changed-file lists
- Snapshot comparison

All inputs normalize to `ChangeScope` with typed `ChangedFile` entries.

## Confidence model

Every impact mapping carries confidence:
- **exact**: direct coverage lineage proves the relationship
- **inferred**: structural heuristics (same directory, naming conventions)
- **weak**: best-effort fallback (no lineage, no structural match)

## Protection status

Code units are classified:
- **strong**: unit or integration test coverage
- **partial**: some coverage but gaps
- **weak**: only E2E or indirect coverage
- **none**: no observed coverage

## Package structure

```
internal/impact/
  impact.go        — models and Analyze() entry point
  analysis.go      — mapping, gap detection, test selection, posture
  changescope.go   — git diff and path-based change scope construction
  impact_test.go   — unit and integration tests
```
