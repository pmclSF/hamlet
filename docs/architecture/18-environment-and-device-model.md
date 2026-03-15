# Environment and Device Model

> **Status:** Implemented
> **Purpose:** Define how Terrain models execution environments and device targets as first-class graph nodes, enabling environment-aware coverage analysis and cross-platform risk detection.
> **Key decisions:**
> - Environments and devices are graph nodes, not metadata annotations on test files
> - Environment coverage is a first-class dimension alongside source-level coverage
> - A test that passes on one environment does not automatically cover another — coverage is per-environment
> - Terrain infers environments from CI configuration files (zero-config principle)
> - Conservative under uncertainty: if environment information cannot be inferred, Terrain reports "unknown environment" rather than assuming universal coverage

**See also:** [02-graph-schema.md](02-graph-schema.md), [16-unified-graph-schema.md](16-unified-graph-schema.md), [12-risk-and-coverage-taxonomy.md](12-risk-and-coverage-taxonomy.md), [26-device-matrix-foundation.md](26-device-matrix-foundation.md)

## Problem

Test results depend on where they run. A test that passes on Linux may fail on macOS. A test that passes on Chrome may break on Safari. A test that passes in CI may fail in staging due to network configuration differences.

Terrain's current model treats test coverage as environment-independent: if a source file has three tests covering it, it gets a "High" coverage band regardless of whether those tests all run on the same platform. This creates blind spots:

- A cross-platform library tested only on Linux has coverage gaps on Windows and macOS that are invisible to current analysis.
- A web application tested only in Chrome has browser-specific risk that current insights cannot surface.
- A mobile app tested only on the latest iOS version has device-specific risk for older versions.

Without environment awareness, Terrain cannot answer questions like: "Does our payment flow have test coverage on Safari?" or "Which tests have never run in staging?"

## Environment Node

An Environment node represents an execution context where tests run.

```
Node Type: Environment
ID Prefix: env:
ID Format: env:<canonical-name>
```

### Metadata

| Field | Type | Description |
|-------|------|-------------|
| `os` | string | Operating system (linux, macos, windows) |
| `osVersion` | string | OS version if known |
| `runtime` | string | Language runtime (node-22, go-1.22, python-3.12) |
| `ciProvider` | string | CI system (github-actions, gitlab-ci, jenkins) |
| `resourceClass` | string | Compute tier if known (large, xlarge, gpu) |
| `isProduction` | boolean | Whether this is a production-like environment |

### Examples

- `env:ci-linux-node22` — GitHub Actions runner on Ubuntu with Node 22
- `env:ci-macos-node22` — GitHub Actions runner on macOS with Node 22
- `env:staging` — Staging environment for integration tests
- `env:local-macos` — Developer local machine on macOS

## Device Node

A Device node represents a target device or browser where tests execute.

```
Node Type: Device
ID Prefix: device:
ID Format: device:<canonical-name>
```

### Metadata

| Field | Type | Description |
|-------|------|-------------|
| `platform` | string | Platform category (ios, android, web-browser) |
| `formFactor` | string | Physical form (phone, tablet, desktop) |
| `osVersion` | string | Device OS version |
| `capabilities` | []string | Device-specific capabilities (touch, camera, biometrics) |
| `browserEngine` | string | Rendering engine for web browsers (chromium, webkit, gecko) |

### Examples

- `device:iphone-15-ios17` — iPhone 15 running iOS 17
- `device:pixel-8-android14` — Pixel 8 running Android 14
- `device:chrome-120` — Chrome 120 desktop browser
- `device:safari-17` — Safari 17 desktop browser

## Edge Types

Two new edge types connect tests to their execution contexts:

### `VALIDATED_IN_ENVIRONMENT`

```
Direction: TestFile → Environment
Confidence: Inferred from CI config analysis (typically 0.8-0.9)
Evidence: ci_config_matrix, ci_config_runs_on, manual_annotation
```

This edge means "this test file is known to execute in this environment." The edge is created when Terrain detects that a CI workflow runs a test suite in a specific environment configuration.

### `TARGETS_DEVICE`

```
Direction: TestFile → Device
Confidence: Inferred from test framework config or browser launch calls (typically 0.7-0.9)
Evidence: framework_config, code_pattern, manual_annotation
```

This edge means "this test file targets this device or browser." The edge is created when Terrain detects browser or device configuration in test setup, Playwright device descriptors, or BrowserStack/Sauce Labs config.

## Impact on Coverage

Environment-aware coverage extends the existing coverage model from [12-risk-and-coverage-taxonomy.md](12-risk-and-coverage-taxonomy.md):

**Current model:** A source file's coverage band is determined by how many tests cover it (High: 3+, Medium: 1-2, Low: 0).

**Extended model:** Coverage bands become environment-qualified. A source file with three tests — all running only on `env:ci-linux-node22` — has High coverage on Linux and Low coverage on macOS and Windows. The overall coverage band reflects the weakest environment when cross-platform support is expected.

Terrain determines whether cross-platform coverage is expected by examining the project's declared platform targets (CI matrix, package.json `os` field, mobile platform configs). If a project only targets Linux, single-environment coverage is not penalized.

## Impact on Confidence

Environment coverage feeds into the confidence model from [14-evidence-scoring-and-confidence-model.md](14-evidence-scoring-and-confidence-model.md):

- A test with `VALIDATED_IN_ENVIRONMENT` edges to multiple environments has higher confidence for cross-platform claims.
- A test with no environment edges has reduced confidence — Terrain cannot verify where it runs.
- Environment inference from CI configs receives confidence 0.8 (CI configs are reliable but may not reflect all execution contexts). Manual annotations receive confidence 0.95.

## Impact on Insights

Environment awareness enables new insight types:

- **Environment gap:** "Your auth tests only run on Chrome — Safari and Firefox are uncovered."
- **Platform blind spot:** "This cross-platform library has no test coverage on Windows."
- **Environment drift:** "Tests pass on CI but staging environment has no recorded test execution."
- **Device concentration:** "90% of your mobile tests target only the latest iOS — older versions are uncovered."

## Inference

Terrain infers environment and device information from CI configuration files, following the zero-config principle. No user annotation is required for common patterns.

### CI Configuration Sources

| Source | Inferred Information |
|--------|---------------------|
| GitHub Actions `runs-on` | OS and runner type |
| GitHub Actions `matrix` | OS, runtime version, browser combinations |
| `.docker` / `Dockerfile` | Base OS, runtime version |
| BrowserStack config | Target browsers and devices |
| Playwright `projects` config | Target browsers |
| Xcode scheme / `destination` | iOS device and version |
| `tox.ini` / `nox` sessions | Python version matrix |

### Inference Confidence

Environment inference is conservative. When Terrain cannot determine the full environment specification, it creates a partial Environment node with reduced confidence rather than guessing. The `explain` command surfaces exactly what was inferred and from which source file, maintaining explainability over convenience.

## Implementation

| File | Purpose |
|------|---------|
| `internal/models/environment.go` | Model structs: `Environment`, `EnvironmentClass`, `DeviceConfig` |
| `internal/models/snapshot.go` | `Environments`, `EnvironmentClasses`, `DeviceConfigs` arrays on `TestSuiteSnapshot` |
| `internal/depgraph/node.go` | Node types: `NodeEnvironment`, `NodeEnvironmentClass`, `NodeDeviceConfig` |
| `internal/depgraph/edge.go` | Edge types: `EdgeTargetsEnvironment`, `EdgeEnvironmentClassContains` |
| `internal/depgraph/build.go` | `buildEnvironments()`, `buildEnvironmentClasses()`, `buildDeviceConfigs()`, `buildEnvironmentEdges()` |
| `internal/depgraph/build_environment_test.go` | 17 tests covering node creation, class membership, deduplication, environment edges |
| `internal/models/test_file.go` | `EnvironmentIDs`, `DeviceIDs` fields on `TestFile` |
| `internal/models/validation_target.go` | `EnvironmentIDs` field on `Scenario` |
| `internal/matrix/matrix.go` | Matrix analysis: `Analyze(g)`, `RecommendForTests(g, paths)` |
| `internal/matrix/matrix_test.go` | 14 tests covering gaps, concentration, recommendations, determinism |

### ID Conventions

| Entity | ID Format | Examples |
|--------|-----------|---------|
| Environment | `env:<canonical-name>` | `env:ci-linux-node22`, `env:staging` |
| EnvironmentClass | `envclass:<name>` | `envclass:browser`, `envclass:os` |
| DeviceConfig | `device:<canonical-name>` | `device:iphone-15-ios17`, `device:chrome-120` |

### Graph Connectivity

- `EnvironmentClass → Environment` via `EdgeEnvironmentClassContains` (membership can be declared on either side)
- `EnvironmentClass → DeviceConfig` via `EdgeEnvironmentClassContains` (devices can belong to classes)
- `TestFile → Environment` via `EdgeTargetsEnvironment` (from `EnvironmentIDs` on TestFile)
- `TestFile → DeviceConfig` via `EdgeTargetsEnvironment` (from `DeviceIDs` on TestFile)
- `Scenario → Environment` via `EdgeTargetsEnvironment` (from `EnvironmentIDs` on Scenario)
- Missing member references are silently skipped — no dangling edges
