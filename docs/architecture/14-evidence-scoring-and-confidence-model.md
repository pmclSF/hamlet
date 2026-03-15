# Evidence Scoring and Confidence Model

> **Status:** Implemented
> **Purpose:** Define how edge confidence scores, hop decay, and fanout penalties combine to produce explainable confidence values for all engine results.
> **Key decisions:**
> - Every graph edge carries a confidence score (0.0-1.0) and an evidence type
> - Path confidence compounds multiplicatively: edge confidences multiplied by 0.85^pathLength decay
> - High-fanout nodes apply a proportional penalty (threshold/fanout) to prevent noisy impact results
> - Minimum confidence threshold of 0.1 excludes likely false positives from results

**See also:** [02-graph-schema.md](02-graph-schema.md), [13-graph-traversal-algorithms.md](13-graph-traversal-algorithms.md), [15-edge-case-handling.md](15-edge-case-handling.md)

## Overview

Every edge in the dependency graph carries a confidence score (0.0 to 1.0) and an evidence type. These scores propagate through the graph to produce confidence values for insight engine results.

## Evidence Types

| Evidence Type | Description | Typical Confidence |
|---------------|-------------|-------------------|
| `static-analysis` | Discovered through AST parsing or regex pattern matching | 0.9 - 1.0 |
| `import-graph` | Discovered through import/require statement resolution | 0.85 - 1.0 |
| `heuristic` | Inferred from naming conventions, directory structure, or co-location | 0.5 - 0.8 |
| `explicit` | Declared in configuration (terrain.yaml, test annotations) | 1.0 |

## Edge Confidence

Edge confidence reflects how certain Terrain is that the relationship exists:

- **1.0** — deterministic. An import statement that resolves to a known file.
- **0.85-0.95** — high confidence. Static analysis found a clear pattern (e.g., fixture import following known conventions).
- **0.5-0.8** — moderate confidence. Heuristic detection (e.g., a file in a `fixtures/` directory is assumed to be a fixture based on location, not import analysis).
- **< 0.5** — low confidence. Speculative or inferred relationships.

## Path Confidence

When traversing the graph, confidence compounds across edges:

```
pathConfidence = edge1.confidence * edge2.confidence * ... * edgeN.confidence
```

Additionally, each hop applies a decay factor of 0.85:

```
pathConfidence = product(edge.confidence) * 0.85^pathLength
```

This ensures that:
- Direct dependencies have high confidence
- Transitive dependencies through long chains have reduced confidence
- The reduction is predictable and explainable

## Fanout Penalty

High-fanout nodes dilute impact signal. When a node has transitive fanout exceeding the threshold (default: 10), confidence is further reduced:

```
fanoutPenalty = threshold / transitiveFanout  (capped at 1.0)
```

For example, a fixture with fanout of 50 and threshold of 10 applies a 0.2 penalty to paths through it.

## Confidence in Engine Results

### Impact Analysis

Each impacted test receives a confidence score reflecting:
- The shortest path from the changed file to the test
- Edge confidences along that path
- Hop decay
- Fanout penalties for intermediate nodes

Tests below a minimum confidence threshold (default: 0.1) are excluded from results.

### Coverage Analysis

Coverage bands (High, Medium, Low) are based on test count, not confidence. However, coverage confidence at the repository level considers graph density — a sparse graph means many relationships may be unobserved.

### Duplicate Detection

Duplicate similarity scores (0.0 to 1.0) are independent of edge confidence. They measure structural similarity between tests based on shared fixtures, helpers, suite paths, and assertion patterns.

## Confidence Adjustments (Edge Cases)

The edge case handler can reduce confidence globally based on repository conditions:

| Edge Case | Adjustment |
|-----------|------------|
| HIGH_SKIP_BURDEN | 0.8x |
| HIGH_FANOUT_FIXTURE | 0.7x |
| LOW_GRAPH_VISIBILITY | 0.6x |
| LARGE_MANUAL_TEST_SUITE | 0.9x |
| EXTERNAL_SERVICE_HEAVY | 0.8x |

Adjustments are multiplicative. A repository with both HIGH_SKIP_BURDEN and LOW_GRAPH_VISIBILITY receives a 0.48x confidence adjustment (0.8 * 0.6).

## Design Rationale

The confidence model is intentionally conservative:

- **Decay penalizes long paths.** A test 5 hops away from a change is less certainly impacted than one 1 hop away.
- **Fanout reduces signal.** When everything depends on everything, impact analysis becomes meaningless. The penalty reflects this.
- **Compounding edge cases reduce confidence.** Multiple repository problems compound uncertainty. The model reflects this rather than hiding it.
- **Minimum thresholds exclude noise.** Tests below 0.1 confidence are more likely false positives than real impact.

The goal is not to produce a single "right" confidence number, but to make the uncertainty visible and explainable.
