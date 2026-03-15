# Test Lifecycle Model and Continuity Inference

## Overview

The test lifecycle model (`internal/lifecycle/`) provides formal classification of how tests evolve across snapshots. While Terrain's identity system (`internal/identity/`) assigns deterministic IDs to tests based on their canonical identity (path, suite hierarchy, name), tests frequently change in ways that alter their identity — renames, file moves, splits, and merges. The lifecycle model bridges this gap by inferring continuity relationships between tests whose IDs differ across snapshots.

## Continuity Classes

Every pair of (from-snapshot, to-snapshot) test cases receives one of these classifications:

| Class | Meaning | Confidence Basis |
|-------|---------|-----------------|
| `exact_continuity` | Same TestID in both snapshots | Deterministic (1.0) |
| `likely_rename` | Same file, similar name | Heuristic (0.4-1.0) |
| `likely_move` | Different file, same or similar name | Heuristic (0.4-1.0) |
| `likely_split` | One old test became multiple new tests | Heuristic (0.6) |
| `likely_merge` | Multiple old tests became one new test | Heuristic (0.6) |
| `removed` | Test present in old snapshot only | Deterministic (1.0) |
| `added` | Test present in new snapshot only | Deterministic (1.0) |
| `ambiguous` | Some similarity but no clear classification | Heuristic (variable) |

## Inference Algorithm

The inference runs in three phases:

### Phase 1: Exact Matching

Tests with identical TestIDs across snapshots are classified as `exact_continuity`. This is the highest-confidence classification and requires no heuristic reasoning. These matches are extracted first and removed from further consideration.

### Phase 2: Heuristic Matching

Remaining unmatched tests are scored pairwise using multiple similarity dimensions:

- **Name similarity** (weight: up to 0.4) — LCS-based string similarity on test names
- **Suite hierarchy similarity** (weight: up to 0.2) — similarity of the describe/suite chain
- **Path similarity** (weight: up to 0.2) — weighted combination of directory (40%) and filename (60%) similarity
- **Canonical identity similarity** (weight: up to 0.2) — overall similarity of the full canonical string

Pairs scoring below 0.4 are discarded. Remaining candidates are sorted by score descending and greedily matched 1:1 (each test participates in at most one heuristic match).

The continuity class is determined by what changed:
- Same file + similar name = rename
- Different file + same name = move
- Different file + similar name = move (different directory) or rename (same directory)
- Moderate similarity on both = ambiguous (or rename if name similarity >= 0.7)

### Phase 3: Split/Merge Detection

After greedy 1:1 matching, remaining unmatched tests are checked for 1:N (split) and N:1 (merge) patterns:

- **Split**: One old test name is a prefix of multiple new test names in the same file/directory
- **Merge**: Multiple old test names share a prefix matching one new test name in the same file/directory

Both require at least 2 targets/sources to trigger (a 1:1 prefix match is handled by Phase 2).

### Phase 4: Residual Classification

Any tests still unmatched after all phases are classified as `removed` (old only) or `added` (new only) with confidence 1.0.

## Evidence Basis

Each mapping carries an `Evidence` list documenting which signals supported the classification:

| Evidence | Meaning |
|----------|---------|
| `exact_id_match` | TestIDs are identical |
| `canonical_similarity` | Canonical identity strings are similar (>= 0.7) |
| `suite_hierarchy_match` | Suite hierarchy strings are similar (>= 0.5) |
| `path_similarity` | File paths are similar (>= 0.5) |
| `coverage_continuity` | Coverage patterns suggest continuity (reserved for future use) |
| `name_similarity` | Test names are similar (>= 0.5) |

## Confidence Scoring

- **1.0**: Exact matches, removed tests, added tests — no ambiguity
- **0.6-1.0**: Heuristic matches — the score equals the similarity score (capped at 1.0)
- **0.6**: Split and merge patterns — fixed confidence reflecting the inherent uncertainty
- **0.4-0.6**: Weak matches — above threshold but low confidence

## Integration with Comparison

The `SnapshotComparison` struct in `internal/comparison/` includes a `LifecycleContinuity` field that is populated automatically when `Compare()` is called. This provides lifecycle context alongside signal deltas, risk changes, and other comparison outputs.

The lifecycle result complements the existing `TestCaseDeltas` (which only tracks added/removed/stable by ID) by providing richer classification of what happened to tests whose IDs changed.

## String Similarity

Similarity is computed via the Longest Common Subsequence (LCS) algorithm, normalized by the length of the longer string. This approach:

- Handles insertions and deletions gracefully (e.g., "should login" vs "should login successfully")
- Is case-insensitive for comparison purposes
- Uses O(n) memory via the two-row DP optimization

Path similarity weights filename (60%) more heavily than directory (40%), reflecting that filename is usually more identity-bearing.

## Known Limitations

1. **No semantic analysis**: The model uses syntactic similarity only. Two tests with identical logic but completely different names will not be matched.

2. **Greedy matching**: The 1:1 matching is greedy (best-first), not optimal. In rare cases, a globally better assignment exists but is missed because a high-scoring pair consumed one of the tests.

3. **Split/merge requires prefix**: Split and merge detection relies on name prefix matching. If a test is split into subtests with unrelated names, it will not be detected.

4. **No cross-file split/merge**: Split and merge detection requires tests to share the same file or directory. Cross-package splits are classified as added+removed.

5. **Coverage continuity not yet implemented**: The `coverage_continuity` evidence basis is defined but not yet used. Future work could use coverage overlap to strengthen continuity inference.

6. **Performance**: The pairwise scoring is O(n*m) where n and m are the counts of unmatched tests. For very large test suites with many changes, this could become expensive. In practice, most tests match exactly and the unmatched sets are small.
