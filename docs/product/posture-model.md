# Posture Model

## What Is Posture?

Posture is Hamlet's answer to "how is our test suite doing?" — not as a single score, but as a set of meaningful dimensions that map directly to action.

Each dimension answers a specific question:

| Dimension | Question |
|-----------|----------|
| **Health** | Are our tests themselves reliable? |
| **Coverage Depth** | Is our exported code actually tested? |
| **Coverage Diversity** | Are we testing with the right mix of test types? |
| **Structural Risk** | Are we blocked from modernizing? |
| **Operational Risk** | Are we following our own rules? |

## Posture Bands

Each dimension resolves to a band:

| Band | What It Means | User Action |
|------|---------------|-------------|
| **Strong** | No significant issues | Maintain current practices |
| **Moderate** | Room for improvement | Plan incremental improvements |
| **Weak** | Needs attention | Prioritize in upcoming work |
| **Elevated** | Widespread problems | Escalate and allocate time |
| **Critical** | Immediate attention needed | Stop and fix before continuing |

## Why Not a Single Score?

A single "test health score of 72" is:
- **Opaque** — what does 72 mean?
- **Not actionable** — what do you fix to get to 80?
- **Fragile** — small changes in weighting change the number

Posture dimensions are:
- **Meaningful** — "coverage depth is weak" means something
- **Actionable** — "weak because 40% of exports are untested" tells you what to do
- **Stable** — dimension semantics don't change between versions

## How Posture Is Computed

Each dimension is backed by a small set of **measurements** — concrete, evidence-based computations that assess specific aspects. For example, health includes:

- Flaky test share
- Skip density
- Dead test share
- Slow test share

Each measurement produces a value, a band, and evidence metadata. The dimension posture is derived from the worst band among its measurements, with adjustments for evidence quality and concentration.

## Evidence and Honesty

Every measurement carries evidence strength:

| Evidence | Meaning |
|----------|---------|
| **Strong** | Direct observation, high confidence |
| **Partial** | Some data, gaps noted |
| **Weak** | Limited data, best-effort |
| **None** | No data for this measurement |

A "strong" posture with no evidence is flagged differently from a "strong" posture with rich runtime data. Hamlet never pretends to know more than it does.

## Posture in Practice

### Summary View

```
Posture: MODERATE
  Health:              STRONG
  Coverage Depth:      WEAK
  Coverage Diversity:  MODERATE
  Structural Risk:     STRONG
  Operational Risk:    STRONG
```

### Drill-Down

`hamlet posture` shows the full evidence:

```
COVERAGE_DEPTH
  Posture: WEAK
  coverage_depth posture is weak. Driven by: coverage_depth.uncovered_exports.

  Measurements:
    coverage_depth.uncovered_exports         37.5% [weak]
      Evidence: partial
      15 of 40 exported code unit(s) appear untested (38%).
      * Test linkage is heuristic-based; some coverage may exist but not be detected.
```

### Trends

`hamlet compare` shows how posture changed:

```
Posture Changes
  coverage_depth             MODERATE → WEAK

Measurement Changes
  coverage_depth.uncovered_exports     +12.5%
    band: moderate → weak
```

## Posture Contracts

Each dimension has contracts that define:

1. **Intended semantics** — what the dimension measures and what it excludes
2. **Evidence expectations** — what data quality is needed for confident assessment
3. **Example findings** — concrete scenarios that influence the dimension

These contracts are documented in [Posture Dimensions](../engineering/posture-dimensions.md).

## Benchmark Safety

In benchmark exports, posture is reduced to dimension → band mappings only. No explanations, file paths, or measurement details are included. This makes posture safe for anonymous comparison across repositories.
