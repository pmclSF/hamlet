// Package reasoning provides shared reasoning primitives for Terrain's
// analysis engines.
//
// These primitives extract common patterns from impact, coverage, redundancy,
// stability, and environment analysis into reusable, composable functions
// that operate over the dependency graph.
//
// Design principles:
//   - Pure functions over graph data — no mutation, no side effects
//   - Confidence and evidence are first-class in all scoring
//   - Scale-safe: all traversals have depth/size caps with graceful degradation
//   - Composable: primitives can be combined for domain-specific analysis
//   - Explainable: all results carry reason chains and provenance
//
// Core primitives:
//   - Reachability: BFS/reverse-BFS with configurable decay and stop conditions
//   - Path scoring: confidence propagation with edge weights, length decay, fanout penalty
//   - Coverage aggregation: reverse-edge coverage counting with band classification
//   - Redundancy candidates: blocking-key generation and similarity scoring
//   - Stability hooks: observation aggregation and trend classification
//   - Fallback expansion: graceful widening when primary reasoning has insufficient evidence
package reasoning
