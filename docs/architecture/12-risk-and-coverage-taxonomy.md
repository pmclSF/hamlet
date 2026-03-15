# Risk and Coverage Taxonomy

> **Status:** Implemented
> **Purpose:** Define the risk dimensions and coverage bands used to score and classify repository health.
> **Key decisions:**
> - Risk scored across three orthogonal dimensions: reliability, change, and speed
> - Coverage bands based on test count per source file (not line-level instrumentation)
> - Coverage confidence accounts for graph density — sparse graphs reduce confidence even when observed coverage looks good
> - Manual coverage modeled as non-executable graph nodes that supplement but never replace automated coverage

**See also:** [14-evidence-scoring-and-confidence-model.md](14-evidence-scoring-and-confidence-model.md), [15-edge-case-handling.md](15-edge-case-handling.md), [05-insight-engine-framework.md](05-insight-engine-framework.md)

## Risk Dimensions

Terrain scores risk across three dimensions. Each dimension aggregates relevant signals into an explainable risk surface.

### Reliability Risk

How likely is the test suite to produce false results?

Contributing signals:
- Flaky tests — tests that pass and fail non-deterministically
- Unstable suites — suites with inconsistent pass rates
- Dead tests — tests that always pass or are never executed
- Skip burden — accumulated skipped tests reducing effective coverage

### Change Risk

How risky is it to make changes to this codebase?

Contributing signals:
- Migration blockers — patterns that prevent framework migration
- Deprecated test patterns — usage of deprecated APIs or patterns
- Framework fragmentation — multiple frameworks increasing maintenance cost
- High-fanout dependencies — changes propagate unpredictably

### Speed Risk

How much does test execution time constrain development?

Contributing signals:
- Slow tests — tests exceeding runtime thresholds
- Runtime budget violations — total suite duration exceeding policy limits
- Large test volume — raw test count contributing to CI pressure

## Risk Levels

Each dimension is scored as: **low**, **medium**, or **high**.

Risk is computed at multiple granularities:
- Repository level
- Package level
- Directory level
- Owner level (using CODEOWNERS or configured ownership)

## Coverage Taxonomy

### Coverage Bands

The coverage engine classifies each source file into a band based on the number of tests that exercise it:

| Band | Test Count | Meaning |
|------|-----------|---------|
| **High** | 3+ tests | Well-covered — multiple tests exercise this code |
| **Medium** | 1-2 tests | Partially covered — some test coverage exists |
| **Low** | 0 tests | Uncovered — no tests directly or indirectly cover this code |

### Direct vs. Indirect Coverage

- **Direct coverage** — a test file explicitly imports the source file
- **Indirect coverage** — a test covers the source through a fixture, helper, or transitive import chain

Both count toward coverage, but indirect coverage contributes lower confidence.

### Coverage Confidence

Coverage confidence is a repository-level metric computed from:

1. **Low band ratio** — what fraction of source files have zero test coverage
2. **Graph density** — how complete the dependency graph is

| Low Band Ratio | Graph Density | Confidence |
|---------------|---------------|------------|
| > 60% | Any | Low |
| 30-60% | Any | Medium |
| < 30% | < 0.5 | Medium |
| < 30% | >= 0.5 | High |

A sparse graph (low density) means many dependency relationships may be missing. Even if observed coverage looks good, the true coverage picture may be incomplete.

### Manual Coverage

Manual coverage entries (from `terrain.yaml`) represent validation that happens outside CI:

- TestRail regression suites
- QA checklists
- Release validation workflows
- Jira verification processes

Manual coverage is modeled as `ManualCoverageArtifact` nodes in the graph. They are not executable and do not contribute to automated coverage counts, but they influence risk reporting by providing additional coverage context for areas that appear uncovered in CI.

### Manual Coverage Presence

| Level | Criteria | Effect |
|-------|----------|--------|
| None | No manual coverage entries | No adjustment |
| Partial | 1-4 entries, or < 3 high-criticality | Minor confidence supplement |
| Substantial | 5+ entries, or 3+ high-criticality | Coverage gaps partially offset in risk reports |
