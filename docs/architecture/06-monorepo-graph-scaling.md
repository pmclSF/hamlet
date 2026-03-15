# Monorepo Graph Scaling

> **Status:** Implemented — basic scaling; package-scoped subgraphs and threshold cutoffs are planned
> **Purpose:** Define strategies for scaling the dependency graph to large monorepos without sacrificing analysis speed or usefulness
> **Key decisions:**
> - Package-scoped subgraphs isolate analysis to a single package, expanding cross-package only when needed
> - Incremental builds avoid full graph reconstruction on every change
> - Fanout detection and confidence decay manage high-fanout node explosion
> - Mixed test cultures are handled as metadata on graph nodes, not separate graph instances

See also: [03-graph-storage-incremental-updates.md](03-graph-storage-incremental-updates.md), [15-edge-case-handling.md](15-edge-case-handling.md)

## Challenge

Monorepos can contain thousands of test files across dozens of packages. The dependency graph must scale without becoming too slow to build or too dense to analyze usefully.

## Package-Scoped Subgraphs

In a monorepo, Terrain can scope graph construction to a package boundary:

- Build the graph for a single package first
- Add cross-package edges only for explicit dependencies
- Use `FILE_BELONGS_TO_PACKAGE` and `PACKAGE_DEPENDS_ON_PACKAGE` edges to model package relationships

This allows engines to operate on a package subgraph for fast local analysis, then expand to cross-package scope when needed.

## Graph Size Considerations

| Repository size | Approximate nodes | Build time target |
|----------------|-------------------|-------------------|
| Small (< 100 tests) | < 500 | < 1s |
| Medium (100-1000 tests) | 500-5000 | < 5s |
| Large (1000-10000 tests) | 5000-50000 | < 30s |
| Very large (10000+ tests) | 50000+ | Use incremental |

## Incremental Builds

For large repositories, full rebuilds are avoided through incremental updates (see [03-graph-storage-incremental-updates.md](03-graph-storage-incremental-updates.md)). Only files changed since the last build are re-analyzed.

## Fanout Management

High-fanout nodes are the primary scaling concern. A single fixture that imports 50 source files creates 50 edges. If 100 tests use that fixture, the transitive fanout is 5000 paths.

Strategies:
- **Fanout detection** flags these nodes so teams can address them
- **Confidence decay** reduces the impact signal for paths through high-fanout nodes
- **Threshold cutoffs** prevent unbounded traversal in engines

## Mixed Test Cultures

Monorepos often contain multiple test frameworks (Jest in one package, Vitest in another, Playwright for E2E). The graph handles this naturally — each test file's framework is metadata on the TestFile node. The `MIXED_TEST_CULTURES` edge case detector flags repos with 3+ frameworks to adjust analysis accordingly.
