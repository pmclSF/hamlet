# Architecture

Hamlet V3 is built around a signal-first architecture.

## Core idea

Raw repository facts and runtime artifacts are not the product.
The product is the structured interpretation of those facts.

That interpretation is expressed as:
- signals
- risk surfaces
- snapshots

## Architecture Layers

### 1. Static Analysis Layer

Purpose:
Understand repository structure without executing tests.

Responsibilities:
- discover test files
- detect frameworks
- extract code units
- link code units to tests
- compute static features for quality/migration analysis

Primary packages:
- internal/analysis
- internal/models

### 2. Runtime Ingestion Layer

Purpose:
Interpret CI and runner artifacts.

Responsibilities:
- parse JUnit XML
- parse Jest/Vitest/other JSON artifacts
- parse coverage reports
- normalize runtime metrics
- detect pass/fail/retry/runtime trends for local snapshot use

Primary packages:
- internal/runtime

### 3. Signal Engine

Purpose:
Transform facts into structured findings.

Responsibilities:
- signal registry
- detector interfaces
- health detectors
- quality detectors
- migration detectors
- governance detectors
- severity/confidence assignment
- explanation and suggested-action generation

Primary packages:
- internal/signals
- internal/health
- internal/quality
- internal/migration
- internal/governance

### 4. Risk Engine

Purpose:
Aggregate signals into actionable risk surfaces.

Responsibilities:
- reliability risk
- change risk
- speed risk
- governance risk
- rollups by:
  - file
  - package
  - module/service
  - team/owner
  - repository

Primary packages:
- internal/scoring
- internal/ownership
- internal/policy

### 4b. Policy / Governance Layer

Purpose:
Evaluate repository state against declared local policy and emit governance signals.

Responsibilities:
- load `.hamlet/policy.yaml`
- evaluate policy rules against snapshot state
- emit governance signals (policyViolation, legacyFrameworkUsage, runtimeBudgetExceeded)
- support CI-friendly enforcement via `hamlet policy check`

Primary packages:
- internal/policy (config model + loader)
- internal/governance (evaluation + signal construction)

### 4c. Ownership Layer

Purpose:
Resolve file ownership for grouping and visibility.

Responsibilities:
- load .hamlet/ownership.yaml (explicit config)
- parse CODEOWNERS files
- directory-based fallback
- attach owner labels to test files and signals

Primary packages:
- internal/ownership

### 5. Reporting / Rendering Layer

Purpose:
Render the snapshot, signals, and risk model into user-facing outputs.

Surfaces:
- CLI human-readable output
- JSON output
- review sections (by owner, directory, migration blockers)
- future markdown
- extension view models
- CI annotation payloads

Primary packages:
- internal/reporting
- cmd/hamlet

## Extension Architecture

The extension is intentionally thin.

Flow:
1. extension executes `hamlet analyze --json`
2. CLI returns a structured snapshot
3. extension renders:
   - Overview
   - Health
   - Quality
   - Migration
   - Review

The extension must not re-implement business logic.

## Snapshot-first model

Every core command should be able to produce or consume a `TestSuiteSnapshot`.

This snapshot is the boundary between:
- engine
- reporting
- future hosted aggregation

## Why this architecture

This architecture allows Hamlet to:
- stay repo-native
- support local OSS value
- support future aggregation and benchmarking
- evolve new detectors without reworking the whole system
- support multiple product surfaces with one engine
