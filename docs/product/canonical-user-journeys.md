# Canonical User Journeys

Terrain supports four primary user journeys. These define the product's public interface and structure how documentation, benchmarking, and testing are organized.

## Quick Walkthrough

Run the four journeys in sequence:

```bash
terrain analyze
terrain insights
terrain impact
terrain explain src/auth/login.test.ts
```

Expected outcome by step:
- `analyze` → clear baseline of framework mix, signal profile, and risk surfaces
- `insights` → prioritized improvement actions
- `impact` → impacted tests and confidence for the current diff
- `explain` → evidence chain behind specific findings or recommendations

## Journey Overview

| Journey | Command | User Question | Example Output | Fixture Repo | Snapshot / Golden Test |
|---------|---------|---------------|----------------|--------------|------------------------|
| Understand the test system | `terrain analyze` | What is the state of our test system? | [analyze-report.md](../examples/analyze-report.md) | `tests/fixtures/sample-repo/` | `cmd/terrain/testdata/analyze.golden` |
| Understand a code change | `terrain impact` | What validations matter for this change? | [impact-report.md](../examples/impact-report.md) | `tests/fixtures/sample-repo/` | `cmd/terrain/testdata/impact.golden` |
| Improve the test suite | `terrain insights` | What should we fix in our test system? | [insights-report.md](../examples/insights-report.md) | `tests/fixtures/sample-repo/` | `cmd/terrain/testdata/insights.golden` |
| Explain Terrain's reasoning | `terrain explain` | Why did Terrain make this decision? | [explain-report.md](../examples/explain-report.md) | `tests/fixtures/sample-repo/` | `cmd/terrain/testdata/explain.golden` |

## Journey 1: Understand the Test System

**Command:** `terrain analyze`

**Persona:** Engineering lead onboarding to a new codebase, or a team starting their first Terrain session.

**Trigger:** "We have no idea what our test system looks like."

**Expected output includes:**
- Tests detected (count, frameworks)
- Repository profile (volume, CI pressure, coverage confidence, redundancy, fanout burden)
- Coverage confidence summary (high/medium/low bands)
- Duplicate cluster count
- High-fanout fixture/helper count
- Skipped test burden
- Weak coverage areas
- CI optimization potential
- Top insight / biggest improvement opportunity

**Engines involved:**
- Signal detection pipeline (framework detection, quality signals, health signals)
- Coverage engine (structural reverse coverage)
- Duplicate engine (fingerprinting + similarity clustering)
- Fanout engine (transitive dependency analysis)
- Repository profiling engine
- Risk scoring engine

**Success criteria:** After running `terrain analyze`, a user should be able to answer: "Is our test system healthy? Where are the biggest problems?"

## Journey 2: Understand a Code Change

**Command:** `terrain impact`

**Persona:** Developer opening a PR, or a CI pipeline evaluating a change.

**Trigger:** "I changed auth code — which tests should I worry about?"

**Expected output includes:**
- Changed areas/packages
- Impacted tests count (and total tests for context)
- Coverage confidence for the change scope
- PR risk level
- Top reason categories (direct dependency, fixture dependency, package dependency)
- Insights affecting confidence/risk
- Fallback level used, if any

**Engines involved:**
- Changed-file detection (git diff)
- Import graph traversal
- Impact graph BFS with confidence decay
- Evidence scoring (edge confidence, path length decay, fanout penalty)
- PR risk scoring
- Fallback policy (edge case detection)

**Success criteria:** After running `terrain impact`, a user should know which tests to run and how confident Terrain is in the recommendation.

## Journey 3: Improve the Test Suite

**Command:** `terrain insights`

**Persona:** Tech lead planning a testing improvement sprint.

**Trigger:** "What should we fix first to improve our test system?"

**Expected output includes:**
- Duplicate clusters (count and top cluster detail)
- High-fanout fixtures/helpers (count and worst offender)
- Weak coverage areas (source files with no test coverage)
- Skipped test debt
- Repository profile and edge cases
- Top ranked improvement opportunities with rationale
- Policy recommendations

**Engines involved:**
- Duplicate engine (structural fingerprinting + similarity)
- Fanout engine (transitive BFS reachability)
- Coverage engine (reverse coverage bands)
- Skip/flake signal detection
- Repository profiling + edge case detection
- Executive summary recommendation engine

**Success criteria:** After running `terrain insights`, a user should have a prioritized list of concrete actions to improve test system health.

## Journey 4: Explain Terrain's Reasoning

**Command:** `terrain explain <target>`

**Persona:** Developer who sees a Terrain finding and wants to understand why.

**Trigger:** "Why did Terrain flag this test?" or "Why is this area marked as weak?"

**Expected output includes:**
- Target metadata (test file details, code unit info, or owner summary)
- Dependency paths or reason chains
- Confidence scores and evidence types
- Connected signals and findings
- Navigation hints to related entities

**Supported target types:**
- Test file path (e.g., `test/auth/login.test.js`)
- Test case ID (e.g., `TRN-4A92C8F1`)
- Code unit name or path
- Owner string
- Finding type or index

**Engines involved:**
- Entity lookup (test files, test cases, code units, owners, findings)
- Signal association
- Dependency graph traversal (for structural context)

**Success criteria:** After running `terrain explain`, a user should understand the evidence chain behind any Terrain decision.

## Journey Transitions

Users naturally flow between journeys:

```
analyze → insights → explain → impact
   │          │          │         │
   │          │          │         └─ PR workflow (CI integration)
   │          │          └─ Deep dive into any finding
   │          └─ Planning: what to fix next
   └─ First run: understand the system
```

The CLI help text reflects this flow:

```
Typical flow:
  1. terrain analyze          understand your test system
  2. terrain insights         find what to improve
  3. terrain impact           see what a change affects
  4. terrain explain <target> understand why
```

## Supporting Commands

These commands remain available but are not primary journeys:

| Command | Relationship to Journeys |
|---------|-------------------------|
| `terrain summary` | Subset of `insights` (leadership-oriented view) |
| `terrain focus` | Subset of `insights` (action-first prioritization) |
| `terrain posture` | Deep evidence behind `analyze` posture assessment |
| `terrain show <entity>` | Same as `explain` but requires entity type prefix |
| `terrain debug *` | Raw engine output for development/debugging |

## Debug Commands

Internal engine views accessible via `terrain debug <engine>`:

| Debug Command | Underlying Engine |
|---------------|-------------------|
| `terrain debug graph` | Dependency graph statistics (nodes, edges, density) |
| `terrain debug coverage` | Structural reverse coverage analysis |
| `terrain debug fanout` | High-fanout node detection |
| `terrain debug duplicates` | Duplicate test cluster analysis |
| `terrain debug depgraph` | Full dependency graph analysis (all engines) |

These are intended for development and debugging, not end-user workflows. The legacy `terrain depgraph` command remains as a backward-compatible alias for `terrain debug depgraph`.
