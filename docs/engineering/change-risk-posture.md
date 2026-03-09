# Change-Risk Posture

## Overview

Change-risk posture summarizes the protection and risk profile of a specific code change. It is computed by `computeChangeRiskPosture()` in `internal/impact/analysis.go` and is distinct from repo-wide posture (which assesses the entire test suite).

## Posture Bands

| Band | Meaning |
|------|---------|
| `well_protected` | All impacted units have adequate coverage, low exposure, single owner |
| `partially_protected` | Some gaps exist but the majority of change is covered |
| `weakly_protected` | Significant gaps, many exposed units, or high coordination needs |
| `high_risk` | Most impacted units lack coverage or the change spans many owners |
| `evidence_limited` | Insufficient data to assess risk confidently |

## Dimensions

Change-risk posture is computed across 4 dimensions. The overall band is the worst dimension.

### Protection

Ratio of unprotected units (ProtectionNone or ProtectionWeak) to total impacted units:
- 0% → `well_protected`
- <30% → `partially_protected`
- <60% → `weakly_protected`
- ≥60% → `high_risk`

### Exposure

Count of exported/public code units affected:
- 0 → `well_protected`
- ≤3 → `partially_protected`
- >3 → `weakly_protected`

### Coordination

Number of distinct owners affected:
- ≤1 → `well_protected`
- ≤3 → `partially_protected`
- >3 → `weakly_protected`

### Instability

Combined score of weak coverage and high complexity among impacted units:
- 0 → `well_protected`
- <0.3 → `partially_protected`
- <0.6 → `weakly_protected`
- ≥0.6 → `high_risk`

## Evidence-Limited Band

`isEvidenceLimited()` returns true when:
- No impacted units AND no impacted tests exist, OR
- All impacted units have `ConfidenceWeak`

When evidence is limited, the posture band is overridden to `evidence_limited` regardless of dimension values, with an explanation noting the data gap.

## Distinction from Repo-Wide Posture

| Aspect | Change-Risk Posture | Repo-Wide Posture |
|--------|-------------------|-------------------|
| Scope | One change/PR | Entire test suite |
| Computed by | `impact.Analyze()` | Measurement layer |
| Dimensions | 4 (protection, exposure, coordination, instability) | 5 (health, coverage depth, diversity, structural, operational) |
| Output | Single band + explanation | 18 measurements with bands |
