# Deterministic Test Identity

> **Status:** Implemented
> **Purpose:** Stable, deterministic identifiers for test cases that persist across runs, refactors, and minor code changes.
> **Key decisions:**
> - Composite key format: `test:<filePath>:<lineNumber>:<testName>` balances stability and uniqueness
> - Line-number-based identity chosen over pure content hashing (which breaks on assertion edits) and name-only (which collides on parameterized tests)
> - Collision detection built into the signal engine to flag duplicate IDs

## Problem

Tests need stable identifiers that persist across runs, refactors, and minor code changes. Without deterministic identity, trend tracking, duplicate detection, and snapshot comparison break whenever a test moves or is renamed.

## Identity Format

Each test is identified by a composite key:

```
test:<filePath>:<lineNumber>:<testName>
```

Example:
```
test:src/auth/login.test.ts:15:should validate credentials
```

Suites follow the same pattern:
```
suite:src/auth/login.test.ts:5:authentication
```

## Why This Format

- **File path** anchors the test to a location in the repository
- **Line number** disambiguates tests with identical names in the same file
- **Test name** provides human readability and survives minor line shifts

## Stability Properties

The identity is stable when:
- The test name does not change
- The file path does not change
- The line number does not change significantly

The identity changes when:
- The test is moved to a different file
- The test is renamed
- Significant code is inserted above the test, shifting its line number

## Trade-offs

Line-number-based identity is a pragmatic choice. Pure content hashing would survive line shifts but would change on any assertion edit. Name-only identity would collide for parameterized tests. The composite approach balances stability and uniqueness for most real-world test suites.

## Usage in Engines

- **Impact engine** — identifies which tests are affected by a change
- **Duplicate engine** — uses test IDs as keys in similarity clusters
- **Coverage engine** — maps source files to covering test IDs
- **Explain** — traces dependency paths starting from a test ID
- **Snapshot comparison** — matches tests across snapshots for trend analysis

See [08-test-similarity-structural-fingerprints.md](08-test-similarity-structural-fingerprints.md) for how test IDs are used in structural fingerprinting for duplicate detection.

## Identity Collisions

The signal engine (Go) includes a collision detector that flags cases where the identity scheme produces duplicate IDs. This is rare but can happen with:
- Programmatically generated tests using the same name
- Copy-pasted test blocks at the same line offset in different files with identical relative paths
