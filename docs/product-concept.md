# Product Concept

Hamlet is a developer-first intelligence layer for test suites.

Its goal is not simply to convert tests.
Its goal is to reveal how much confidence the test system actually provides.

## Core Product Model

Hamlet operates across four dimensions.

### Structure
What exists?

Examples:
- frameworks in use
- test file inventory
- ownership layout
- package hotspots
- code-to-test relationships
- framework fragmentation

### Health
What is unstable, slow, or broken?

Examples:
- flaky tests
- slow tests
- skipped tests
- dead tests
- unstable suites
- retry patterns

### Quality
How meaningful are the tests?

Examples:
- untested exports
- weak assertions
- mock-heavy tests
- tests that only verify mock interactions
- snapshot overuse
- coverage blind spots
- broken coverage thresholds

### Change Readiness
How risky is evolution?

Examples:
- framework migration readiness
- migration blockers
- deprecated test patterns
- legacy framework drift
- policy violations
- risk surfaces by module/team

## Product Surfaces

### Local
- CLI
- JSON snapshots
- repo-local state
- VS Code / Cursor extension
- CI annotations

### Hosted / Paid
- trends
- historical snapshots
- org rollups
- benchmarks
- centralized policy
- risk maps across repos

## Product Wedge

Migration remains the entry wedge because it is painful and immediate.

But the broader product becomes:
**observability and intelligence for the health and evolution of the test system**
