# Impact Drill-Down CLI

## Stage 129 -- Drill-Down Views Guide

### Overview

The `terrain impact --show` flag provides six drill-down views into impact analysis results. Each view answers a different question about how a code change relates to the test suite.

---

### Units View (`--show units`)

**Question:** What code units are affected by this change?

```bash
terrain impact --show units
```

**Example output:**

```
Impacted Code Units (14):

  src/auth/tokenValidator.js
    validateAccessToken    [protected]  3 tests (2 exact, 1 inferred)
    validateRefreshToken   [gap]        0 tests
    parseTokenClaims       [protected]  1 test (1 exact)

  src/auth/sessionStore.js
    createSession          [protected]  2 tests (2 exact)
    clearExpired           [gap]        0 tests

  src/middleware/rateLimit.js
    checkRate              [protected]  1 test (1 exact)
    checkBurst             [gap]        0 tests
    resetBucket            [protected]  1 test (1 inferred)
```

Each unit shows its protection status and the number of mapped tests with their confidence levels.

---

### Gaps View (`--show gaps`)

**Question:** Where is the changed code unprotected?

```bash
terrain impact --show gaps
```

**Example output:**

```
Protection Gaps (3 units):

  src/auth/tokenValidator.js :: validateRefreshToken
    Changed at line 84. No mapped tests.
    Nearest candidate: test/auth/tokenValidator.test.js (inferred, low confidence)

  src/auth/sessionStore.js :: clearExpired
    Changed at line 31. No mapped tests.
    No candidate tests found.

  src/middleware/rateLimit.js :: checkBurst
    Changed at line 57. No mapped tests.
    Nearest candidate: test/middleware/rateLimit.test.js (inferred, low confidence)
```

Gaps are the highest-priority items for review. The "nearest candidate" hint suggests where a test could be added or where an existing test might already cover the unit but was not detected.

---

### Tests View (`--show tests`)

**Question:** Which tests exercise the changed code?

```bash
terrain impact --show tests
```

**Example output:**

```
Impacted Tests (11):

  test/auth/tokenValidator.test.js
    Confidence: exact
    Exercises: validateAccessToken, parseTokenClaims
    Frameworks: jest

  test/auth/sessionStore.test.js
    Confidence: exact
    Exercises: createSession
    Frameworks: jest

  test/auth/integration/auth-flow.test.js
    Confidence: inferred
    Exercises: validateAccessToken (transitive via loginHandler)
    Frameworks: jest

  test/middleware/rateLimit.test.js
    Confidence: exact
    Exercises: checkRate, resetBucket
    Frameworks: jest
```

Tests are sorted by confidence (exact first) and grouped by file. The "exercises" field shows which impacted units each test covers.

---

### Owners View (`--show owners`)

**Question:** Which teams or owners are affected?

```bash
terrain impact --show owners
```

**Example output:**

```
Owner Impact:

  @backend-team
    Changed units: 7
    Protective tests: 6
    Gaps: 2
    Posture: MODERATE

  @platform-team
    Changed units: 4
    Protective tests: 3
    Gaps: 1
    Posture: LOW

  (unowned)
    Changed units: 3
    Protective tests: 2
    Gaps: 0
    Posture: LOW
```

Owner mappings come from `.terrain/ownership.yaml`, CODEOWNERS, and optional git-history fallback. Use `--owner` to filter any view to a single owner:

```bash
terrain impact --show units --owner @backend-team
```

---

### Graph View (`--show graph`)

**Question:** How do changes propagate to tests?

```bash
terrain impact --show graph
```

**Example output:**

```
Impact Graph:

  src/auth/tokenValidator.js::validateAccessToken
    <- test/auth/tokenValidator.test.js (exact, direct import)
    <- test/auth/integration/auth-flow.test.js (inferred, transitive)

  src/auth/tokenValidator.js::validateRefreshToken
    (no test edges)

  src/auth/sessionStore.js::createSession
    <- test/auth/sessionStore.test.js (exact, direct import)

  src/middleware/rateLimit.js::checkRate
    <- test/middleware/rateLimit.test.js (exact, direct import)
```

The graph view shows the edges between changed code units and their tests. Units with no test edges are gaps. The edge label indicates confidence and mapping type (direct import, transitive dependency, naming convention).

---

### Selected View (`--show selected`)

**Question:** What tests should I run?

```bash
terrain impact --show selected
```

**Example output:**

```
Selected Protective Test Set (11 tests):

  test/auth/tokenValidator.test.js
  test/auth/sessionStore.test.js
  test/auth/integration/auth-flow.test.js
  test/middleware/rateLimit.test.js
  test/middleware/cors.test.js
  test/api/userRoutes.test.js
  test/api/authRoutes.test.js
  test/api/rateLimitRoutes.test.js
  test/utils/tokenHelpers.test.js
  test/utils/timeUtils.test.js
  test/integration/api-auth.test.js

Run with:
  terrain select-tests --format paths | xargs npx jest
```

This is the same set produced by `terrain select-tests`, presented with context about why each test was selected.

---

### Combining Views

Use `--json` with any view to get structured output for scripting:

```bash
terrain impact --show gaps --json | jq '.gaps[] | .unit'
```

Use `--owner` with any view to scope results:

```bash
terrain impact --show tests --owner @backend-team
```
