# AI User Journey

> **Status:** Implemented
> **Purpose:** Document the end-to-end workflow for AI/ML teams using Terrain to manage eval validation alongside traditional tests.

## The Workflow

```
PR changes a prompt or dataset
    ↓
terrain impact          → identifies impacted tests AND scenarios
    ↓
terrain ai run          → executes impacted eval scenarios (via Gauntlet)
    ↓
terrain explain <id>    → traces why a scenario was selected
```

This is the AI equivalent of the traditional PR workflow (code change → impact → test run → explain). The key difference: **scenarios** are first-class validation targets alongside tests.

## Step-by-Step

### 1. Detect What Changed

A developer modifies a prompt template or dataset loader:

```bash
# See what Terrain detects
terrain impact --base main
```

```
Terrain Impact Analysis
============================================================

Summary: 1 file(s) changed, 2 code unit(s) impacted,
         1 test(s) relevant, 2 scenario(s) impacted.
         Posture: partially_protected.

Changed areas:
  src/ai                 src/ai/prompts.ts (modified)

Impacted tests:          1
Coverage confidence:     Medium

Impacted Scenarios (2)
------------------------------------------------------------
  safety-check (safety) [exact]
    covers 1 changed surface(s)
  accuracy-check (accuracy) [exact]
    covers 2 changed surface(s)

Next steps:
  terrain impact --show selected   view protective test set
  terrain explain <scenario-id>    understand why
```

**What happened:** Terrain detected that `prompts.ts` changed, identified the prompt surfaces inside it, found 2 scenarios that validate those surfaces, and surfaced them alongside the traditional test impact.

### 2. Run Impacted Evaluations

```bash
# Execute the impacted scenarios via Gauntlet
terrain ai run
```

When fully implemented, this command will:
1. Read the impacted scenario list from the most recent impact analysis
2. Invoke Gauntlet as the execution backend
3. Collect results and ingest them back into Terrain

Currently scaffolded — use Gauntlet directly and feed results back:

```bash
# Run Gauntlet manually
gauntlet run --output results.json

# Ingest results
terrain analyze --root . --gauntlet results.json
```

### 3. Explain Why a Scenario Was Selected

```bash
terrain explain scenario:custom:safety-check
```

```
Terrain Explain — Scenario
============================================================

Scenario: safety-check
Category: safety
Framework: promptfoo
Confidence: exact

Verdict: Scenario "safety-check" is impacted because its covered
         surface surface:src/ai/prompts.ts:systemPrompt changed.

Changed surfaces (1):
  surface:src/ai/prompts.ts:systemPrompt

Next steps:
  terrain ai list              view all detected scenarios
  terrain impact --json         machine-readable impact data
```

**What happened:** Terrain traced the impact from the changed file through the code surface to the scenario that declares coverage of that surface, producing an explainable reason chain.

### 4. Review the Full Picture

```bash
# See all AI/eval assets
terrain ai list

# Validate setup
terrain ai doctor

# Full analysis with Gauntlet results
terrain analyze --root . --gauntlet results.json
```

## How Scenarios Get Into Terrain

### Option 1: Declare in `.terrain/terrain.yaml`

```yaml
scenarios:
  - name: safety-check
    category: safety
    framework: promptfoo
    owner: ml-team
    surfaces:
      - "surface:src/ai/prompts.ts:systemPrompt"
      - "surface:src/ai/prompts.ts:userTemplate"
    path: evals/safety.yaml

  - name: accuracy-regression
    category: accuracy
    framework: deepeval
    owner: ml-team
    surfaces:
      - "surface:src/models/classifier.py:classify"
    path: evals/accuracy.py
```

The `surfaces` field links scenarios to code surfaces by ID. When those surfaces change, the scenario is flagged as impacted.

### Option 2: Auto-Detection (Future)

Terrain will detect eval frameworks (deepeval, promptfoo, ragas) from config files and directory conventions, creating scenarios automatically without YAML configuration.

## What Each Command Provides

| Command | AI-Specific Output |
|---------|-------------------|
| `terrain analyze` | Validation inventory: scenarios, prompts, datasets. Scenario duplication in insights. |
| `terrain impact` | `ImpactedScenarios[]` with confidence, relevance, and covered surfaces |
| `terrain explain <scenario>` | Verdict explaining why the scenario is impacted, with changed surface list |
| `terrain ai list` | All detected scenarios, prompt surfaces, dataset surfaces, eval files |
| `terrain ai doctor` | 5-point diagnostic: scenarios, prompts, datasets, eval files, graph wiring |
| `terrain insights` | AI scenario duplication finding when scenarios overlap >50% of surfaces |

## The Reasoning Path

```
CodeSurface           → code function, prompt, or dataset
    ↑
BehaviorSurface       → behavioral grouping (derived)
    ↑
Scenario              → eval case declaring surface coverage
    ↓
Environment           → where the eval runs
    ↓
ExecutionRun          → Gauntlet execution record
```

When a CodeSurface changes:
1. **Impact** finds all Scenarios whose `CoveredSurfaceIDs` include that surface
2. **Explain** traces from the scenario back through the surface to the changed file
3. **Gauntlet** executes the scenario and produces a result artifact
4. **Analyze** ingests the Gauntlet artifact, generating signals for failures/regressions

## Confidence Model

| Relationship | Confidence | Source |
|---|---|---|
| Scenario → CodeSurface | exact | Declared in `terrain.yaml` via `surfaces` field |
| BehaviorSurface → CodeSurface | 0.7 | Inferred from code analysis |
| Test → CodeUnit | exact or inferred | Import graph or structural heuristics |

Scenarios that explicitly declare their covered surfaces get `exact` confidence — higher than inferred test-to-code mappings. This rewards teams that invest in declaring eval coverage targets.

## Integration Tests

The AI workflow is validated by 7 integration tests in `cmd/terrain/ai_workflow_test.go`:

| Test | What It Validates |
|------|-------------------|
| `TestAIWorkflow_PipelineLoadsScenarios` | Pipeline loads 3 scenarios from `.terrain/terrain.yaml` |
| `TestAIWorkflow_ImpactDetectsScenarios` | `terrain impact` surfaces changed areas when classifier changes |
| `TestAIWorkflow_ExplainScenario` | `terrain explain` produces verdict for scenario by ID and name |
| `TestAIWorkflow_ExplainScenario_NilResult` | Graceful error on nil impact result |
| `TestAIWorkflow_AIListShowsScenarios` | `terrain ai list` shows scenarios and dataset surfaces |
| `TestAIWorkflow_AIDoctorPassesWithScenarios` | `terrain ai doctor` passes 4/5 checks with configured scenarios |
| `TestAIWorkflow_FullChain_Deterministic` | Full pipeline produces identical output across runs |

Fixture: `tests/fixtures/ai-eval-suite/` with `.terrain/terrain.yaml` defining 3 scenarios covering classifier, embeddings, and dataset surfaces.
