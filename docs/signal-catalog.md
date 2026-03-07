# Signal Catalog

This file defines the initial V3 signal catalog.

Each signal includes:
- category
- description
- why it matters
- remediation direction

---

## Health Signals

### slowTest
Category: health

Description:
A test or suite consistently exceeds an expected runtime threshold.

Why it matters:
Slow tests create CI bottlenecks, slow feedback loops, and increase migration validation cost.

Suggested remediation direction:
- reduce setup overhead
- isolate expensive integration behavior
- move or split large runtime hotspots

---

### flakyTest
Category: health

Description:
A test demonstrates intermittent failures or elevated retry behavior.

Why it matters:
Flakes reduce trust in the test suite and make changes harder to validate.

Suggested remediation direction:
- identify nondeterministic dependencies
- reduce timing assumptions
- isolate unstable fixtures or mocks

---

### skippedTest
Category: health

Description:
A test is disabled, skipped, or pending.

Why it matters:
Skipped tests create false confidence and often conceal degraded quality.

Suggested remediation direction:
- restore or remove intentionally
- track stale skips
- prevent accumulation in CI

---

### deadTest
Category: health

Description:
A test appears disconnected from live behavior, modules, or execution paths.

Why it matters:
Dead tests increase maintenance cost while providing little or no confidence.

Suggested remediation direction:
- delete if obsolete
- reconnect if intended to remain active
- investigate orphaned snapshots and references

---

### unstableSuite
Category: health

Description:
A suite exhibits unusually high variance, retries, or inconsistency as a group.

Why it matters:
Suite-level instability often indicates shared fixture or infrastructure problems.

Suggested remediation direction:
- inspect common setup/teardown
- isolate environmental dependencies
- reduce shared mutable state

---

## Quality Signals

### untestedExport
Category: quality

Description:
An exported function, method, or public code unit appears to have weak or missing direct test coverage.

Why it matters:
Public APIs with weak coverage create high change risk.

Suggested remediation direction:
- add direct tests
- improve code-to-test linkage
- prioritize frequently changed or critical exports

---

### weakAssertion
Category: quality

Description:
A test file or suite has low or weak assertion strength relative to its scope.

Why it matters:
Code may execute without meaningfully verifying behavior.

Suggested remediation direction:
- assert on outputs, state changes, side effects, or user-visible behavior
- expand edge-case and negative-path checks

---

### mockHeavyTest
Category: quality

Description:
A test relies heavily on mocks relative to real interactions.

Why it matters:
Mock-heavy tests can overstate confidence while missing real integration behavior.

Suggested remediation direction:
- reduce unnecessary mocking
- add assertions on real system behavior
- supplement with integration coverage

---

### testsOnlyMocks
Category: quality

Description:
Assertions primarily or exclusively validate mock interactions rather than business outcomes.

Why it matters:
These tests often verify implementation details rather than meaningful behavior.

Suggested remediation direction:
- assert on returned values
- assert on persisted state, domain events, rendered UI, or side effects

---

### snapshotHeavyTest
Category: quality

Description:
A file depends heavily on snapshots relative to direct semantic assertions.

Why it matters:
Snapshot overuse can hide weak behavioral coverage and create brittle review noise.

Suggested remediation direction:
- replace some snapshots with targeted assertions
- reduce low-value snapshot churn

---

### coverageBlindSpot
Category: quality

Description:
Coverage exists, but high-risk paths or code units remain weakly exercised.

Why it matters:
Raw coverage percentages can hide meaningful quality gaps.

Suggested remediation direction:
- improve branch/path coverage
- focus on high-complexity or critical modules

---

### coverageThresholdBreak
Category: quality

Description:
Coverage is below a declared threshold.

Why it matters:
A threshold break is a concrete signal of degraded test effectiveness or change quality.

Suggested remediation direction:
- identify concentrated gaps
- target high-risk modules first
- distinguish broad threshold issues from critical blind spots

---

## Migration Signals

### frameworkMigration
Category: migration

Description:
The repository or package appears suitable for migration from one framework to another.

Why it matters:
Provides modernization guidance and helps prioritize change.

Suggested remediation direction:
- review representative examples
- estimate blockers before conversion

---

### migrationBlocker
Category: migration

Description:
A pattern makes automated or safe migration difficult.

Why it matters:
Blockers determine manual effort and migration risk.

Suggested remediation direction:
- group by blocker type
- address high-frequency blockers first
- route complex blockers to review

---

### deprecatedTestPattern
Category: migration

Description:
A test pattern is outdated or poorly aligned with target framework or future standards.

Why it matters:
Deprecated patterns increase future migration and maintenance cost.

Suggested remediation direction:
- update patterns early
- include in modernization backlog

---

### dynamicTestGeneration
Category: migration

Description:
Dynamic generation patterns reduce migration predictability.

Why it matters:
These patterns are often hard to translate safely.

Suggested remediation direction:
- isolate generation logic
- review manually
- simplify when possible

---

### customMatcherRisk
Category: migration

Description:
Custom matchers or helper abstractions complicate portability.

Why it matters:
Migration automation is less reliable when semantics are hidden behind local wrappers.

Suggested remediation direction:
- inventory wrappers
- add mapping support or refactor to standard assertions

---

## Governance Signals

### policyViolation
Category: governance

Description:
Current repository state violates declared Hamlet policy.

Why it matters:
Policy violations indicate drift or unmanaged risk.

Suggested remediation direction:
- review local policy configuration
- fix or explicitly waive with rationale

---

### legacyFrameworkUsage
Category: governance

Description:
Legacy or disallowed framework usage persists or is reintroduced.

Why it matters:
This can stall modernization and fragment standards.

Suggested remediation direction:
- prevent new usage
- prioritize migration hotspots
- create path-specific policies if needed

---

### skippedTestsInCI
Category: governance

Description:
Skipped tests are present where CI policy disallows them.

Why it matters:
Enforced visibility prevents silent quality erosion.

Suggested remediation direction:
- remove or restore skipped tests
- use limited exceptions explicitly

---

### runtimeBudgetExceeded
Category: governance

Description:
Tests or suites exceed configured runtime budgets.

Why it matters:
Runtime budgets protect feedback loops and CI costs.

Suggested remediation direction:
- isolate hotspots
- refactor test setup
- adjust policy only with explicit intent
