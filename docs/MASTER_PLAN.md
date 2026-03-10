# Hamlet Master Plan

## Status

This document is the primary source of truth for the current refactor.

Hamlet's current engine is a strategic evolution from a framework migration tool into:

**Observability and intelligence for test suites**

The purpose of this plan is to rebuild the product around a durable signal model, a risk model, and a multi-surface architecture that supports:

- local repository analysis
- test suite health visibility
- test quality intelligence
- migration readiness analysis
- governance and policy checks
- future historical and organizational intelligence

This file defines the product direction, implementation boundaries, architectural priorities, and milestone sequence.

---

## Product Definition

Hamlet is the control plane for test suite intelligence.

Hamlet does not run tests.

Instead, it observes and interprets signals from:

- repository structure
- test files
- code units
- CI artifacts
- coverage reports
- ownership metadata
- local policy/configuration

Hamlet helps teams answer:

- What tests do we have?
- Which parts of the test system are fragile?
- Which tests are flaky, slow, skipped, or dead?
- Which tests provide weak confidence?
- Which parts of the codebase are risky to change?
- How ready are we to migrate or modernize the test stack?
- How do we prevent regression over time?

---

## Product Pillars

### 1. Structure
Understand the shape of the test ecosystem.

Includes:
- frameworks
- test file inventory
- code unit mapping
- ownership
- package/module structure
- framework fragmentation

### 2. Health
Understand reliability and runtime behavior.

Includes:
- flaky tests
- slow tests
- skipped tests
- dead tests
- unstable suites
- retry rates
- runtime hotspots

### 3. Quality
Understand whether tests provide meaningful confidence.

Includes:
- untested exports
- weak assertions
- mock-heavy tests
- tests that only validate mocks
- snapshot overuse
- coverage blind spots
- coverage threshold breaks

### 4. Change Readiness
Understand how risky it is to evolve the codebase.

Includes:
- migration readiness
- migration blockers
- deprecated patterns
- policy violations
- legacy framework usage
- modernization risk

---

## Core Principles

1. Signals are the primary abstraction.
2. Analysis comes before automation.
3. Risk must be explainable.
4. Hamlet must remain repo-native and useful without SaaS.
5. Open source should include the core signals and local experience.
6. Paid value should come from aggregation, history, benchmarking, and governance.
7. Hamlet measures system health, not individual developer productivity.

---

## Canonical UX Flow

Observe -> Understand -> Act -> Improve

This maps to:
- Overview
- Health
- Quality
- Migration
- Review
- Policy

---

## Refactor Goals

### Goal 1
Standardize the signal model and snapshot model.

### Goal 2
Refactor the engine around clean layers:
- static analysis
- runtime ingestion
- signal generation
- risk scoring
- reporting

### Goal 3
Introduce a stable `hamlet analyze` command as the primary entry point.

### Goal 4
Make risk a first-class product output.

### Goal 5
Create a thin VS Code / Cursor extension over CLI JSON output.

### Goal 6
Preserve migration features, but reposition them as one capability within test intelligence.

---

## What This Plan Is Not

This plan is not:
- a test runner
- a CI system
- a generic static analysis platform
- a dashboard-first SaaS product
- a developer surveillance system

---

## Implementation Strategy

This plan will be implemented in vertical slices.

### Slice 1
Documentation, repo structure, models, signal registry

### Slice 2
Static analysis nucleus

### Slice 3
Initial signal engine

### Slice 4
Risk engine

### Slice 5
CLI analyze

### Slice 6
Snapshot persistence and extension scaffold

### Slice 7
Policy/governance checks

---

## Initial Signal Set

### Health
- slowTest
- flakyTest
- skippedTest
- deadTest
- unstableSuite

### Quality
- untestedExport
- weakAssertion
- mockHeavyTest
- testsOnlyMocks
- snapshotHeavyTest
- coverageBlindSpot
- coverageThresholdBreak

### Migration
- frameworkMigration
- migrationBlocker
- deprecatedTestPattern
- dynamicTestGeneration
- customMatcherRisk

### Governance
- policyViolation
- legacyFrameworkUsage
- skippedTestsInCI
- runtimeBudgetExceeded

---

## Risk Dimensions

### Reliability Risk
Flakes, retries, dead tests, skipped tests, unstable suites

### Change Risk
Weak tests, untested exports, low coverage, migration blockers

### Speed Risk
Slow tests, slow suites, CI hotspots

### Governance Risk
Policy violations, drift, budget overflow

---

## Open Source / Paid Boundary

### Open Source
- local analysis
- local signals
- local risk
- CLI
- local JSON snapshots
- extension
- local policy checks
- migration readiness

### Paid
- trends over time
- cross-repo rollups
- org and team comparison
- benchmark comparison to similar organizations
- centralized policy
- org-level risk maps
- risk budgets
- hosted visibility

---

## Final Rule

Every signal and every score must answer:
**What should the user do next?**
