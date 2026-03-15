# Test Similarity and Structural Fingerprints

> **Status:** Implemented
> **Purpose:** Detect redundant tests through structural similarity analysis rather than text comparison
> **Key decisions:**
> - Four weighted signals (fixture overlap, helper overlap, suite path similarity, assertion pattern similarity) compose the similarity score
> - Jaccard similarity is used for set-based overlap calculations
> - Default clustering threshold of 0.6 balances precision and recall for duplicate detection
> - Algorithm is intentionally O(n^2); package-scoped analysis is recommended for very large suites

See also: [05-insight-engine-framework.md](05-insight-engine-framework.md), [04-deterministic-test-identity.md](04-deterministic-test-identity.md)

## Problem

Test suites grow organically. Teams copy existing tests as templates, write overlapping tests across integration and unit layers, and inherit duplicate coverage through shared fixtures. Identifying redundant tests is essential for reducing CI time without reducing coverage.

## Approach: Structural Fingerprinting

Terrain detects duplicates through structural similarity rather than text comparison. Two tests are structurally similar if they exercise the same code paths through the same mechanisms, even if their source code looks different.

## Similarity Signals

Each test pair is scored across four weighted dimensions:

| Signal | Weight | Description |
|--------|--------|-------------|
| Fixture overlap | 25% | Fraction of shared fixtures between two tests |
| Helper overlap | 25% | Fraction of shared helper modules |
| Suite path similarity | 30% | Similarity of the describe/suite nesting hierarchy |
| Assertion pattern similarity | 20% | Similarity of assertion structures and patterns |

### Fixture Overlap

Computed as Jaccard similarity of the fixture sets used by each test:

```
fixtureOverlap = |fixturesA ∩ fixturesB| / |fixturesA ∪ fixturesB|
```

Fixtures are identified by following TEST_USES_FIXTURE edges from the test's file node.

### Helper Overlap

Same Jaccard computation for helper modules (TEST_USES_HELPER edges):

```
helperOverlap = |helpersA ∩ helpersB| / |helpersA ∪ helpersB|
```

### Suite Path Similarity

The suite path is the chain of describe blocks containing a test. For example:

```
authentication > login > should validate credentials
```

Suite paths are compared using normalized string similarity. Tests in the same describe hierarchy score higher.

### Assertion Pattern Similarity

Assertion patterns are extracted from test metadata. Tests using the same types of assertions (equality checks, existence checks, error expectations) in the same sequence score higher.

## Scoring

The weighted score is computed as:

```
score = 0.25 * fixtureOverlap
      + 0.25 * helperOverlap
      + 0.30 * suitePathSimilarity
      + 0.20 * assertionPatternSimilarity
```

## Clustering

Tests with pairwise similarity above the threshold (default: 0.6) are grouped into clusters. Each cluster represents a set of tests that are likely redundant.

The duplicate engine reports:
- Total tests analyzed
- Number of tests in duplicate clusters
- Each cluster with member tests and similarity scores

## Threshold Selection

| Threshold | Behavior |
|-----------|----------|
| 0.8+ | Very conservative — only near-identical tests |
| 0.6 (default) | Balanced — catches most meaningful duplicates |
| 0.4 | Aggressive — may include tests that share infrastructure but test different behavior |

The threshold is configurable via `--threshold` on the `terrain duplicates` command.

## Limitations

- Structural fingerprinting does not analyze test logic. Two tests that use the same fixtures and helpers but assert different behaviors will score as duplicates.
- The algorithm is O(n^2) in the number of tests. For very large suites, consider scoping to a package.
- Parameterized tests may appear as duplicates of their base test.
