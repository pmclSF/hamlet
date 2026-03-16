# Gauntlet Integration

> **Status:** Implemented (artifact ingestion)
> **Repository:** github.com/pmclsf/gauntlet
> **Purpose:** Gauntlet is Terrain's first AI execution provider. It runs eval scenarios deterministically and produces structured result artifacts that Terrain ingests for reasoning, coverage analysis, and baseline comparison.

## Responsibility Split

| Concern | Owner | Implementation |
|---------|-------|---------------|
| Scenario selection | Terrain | `terrain ai list`, graph-based coverage queries |
| Reasoning & explanation | Terrain | Impact analysis, coverage reasoning, `terrain explain` |
| Coverage analysis | Terrain | `AnalyzeCoverage`, posture dimensions, coverage gaps |
| Deterministic execution | Gauntlet | Eval runner with reproducible ordering and environment control |
| Baseline comparison | Gauntlet | Result diffing, regression detection, metric tracking |
| Result artifacts | Gauntlet | JSON output conforming to Terrain's ingestion schema |

Terrain never executes evals. Gauntlet never reasons about coverage or risk. The boundary is the artifact file.

## Artifact Format

Gauntlet produces a JSON artifact that Terrain ingests via `--gauntlet`. The schema:

```json
{
  "version": "1",
  "provider": "gauntlet",
  "timestamp": "2026-03-15T12:00:00Z",
  "repository": "my-app",
  "scenarios": [
    {
      "scenarioId": "eval:safety-check",
      "name": "safety-check",
      "status": "passed",
      "durationMs": 1200,
      "metrics": {
        "accuracy": 0.95,
        "latency_p95_ms": 800
      },
      "modelVersion": "gpt-4o-2024-08-06",
      "baseline": {
        "accuracy": 0.93,
        "latency_p95_ms": 750
      },
      "regressions": []
    }
  ],
  "summary": {
    "total": 10,
    "passed": 8,
    "failed": 1,
    "skipped": 1,
    "durationMs": 15000
  }
}
```

### Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | string | yes | Schema version (`"1"`) |
| `provider` | string | yes | Always `"gauntlet"` |
| `timestamp` | string | yes | ISO 8601 execution timestamp |
| `repository` | string | no | Repository name for cross-reference |
| `scenarios` | array | yes | Per-scenario execution results |
| `scenarios[].scenarioId` | string | yes | Matches Terrain's `Scenario.ScenarioID` |
| `scenarios[].name` | string | yes | Human-readable scenario name |
| `scenarios[].status` | string | yes | `"passed"`, `"failed"`, `"skipped"`, `"error"` |
| `scenarios[].durationMs` | float | no | Execution time in milliseconds |
| `scenarios[].metrics` | object | no | Key-value metric results (accuracy, latency, etc.) |
| `scenarios[].modelVersion` | string | no | Model version under evaluation |
| `scenarios[].baseline` | object | no | Previous baseline metric values |
| `scenarios[].regressions` | array | no | Metric names that regressed vs baseline |
| `summary` | object | yes | Aggregate execution counts |

## Usage

### Basic Ingestion

```bash
# Run Gauntlet to produce results
gauntlet run --output gauntlet-results.json

# Feed results into Terrain analysis
terrain analyze --root . --gauntlet gauntlet-results.json
```

### Combined with Other Artifacts

```bash
terrain analyze --root . \
  --coverage coverage/lcov.info \
  --runtime test-results.xml \
  --gauntlet gauntlet-results.json
```

### AI Workflow

```bash
# 1. See what Terrain knows about your AI validation
terrain ai list

# 2. Run Gauntlet execution
gauntlet run --scenarios terrain-scenarios.json --output results.json

# 3. Ingest results back into Terrain
terrain analyze --root . --gauntlet results.json

# 4. Check impact of a code change on eval coverage
terrain impact --base main
```

## How Terrain Uses Gauntlet Artifacts

When a Gauntlet artifact is ingested:

1. **Scenario matching** — Each `scenarioId` in the artifact is matched to Terrain's Scenario inventory. Matched scenarios receive execution metadata.

2. **Signal generation** — Failed scenarios generate signals:
   - `"failed"` → medium-severity signal with metric details
   - `"error"` → high-severity signal indicating execution failure
   - Regressions → medium-severity signal with baseline comparison

3. **Data completeness** — The snapshot's `DataSources` includes a `"gauntlet"` entry tracking ingestion status.

4. **Coverage enrichment** — Scenario execution results improve coverage confidence for the code surfaces those scenarios validate.

## Graph Integration

Gauntlet results connect through the AI reasoning path:

```
CodeSurface ← Scenario (with Gauntlet execution data)
                  ↓
              Environment (execution target)
                  ↓
              ExecutionRun (Gauntlet run metadata)
```

When `terrain explain <scenario-id>` is called, the explanation traces include Gauntlet execution status, metric values, and baseline comparisons.

## Benchmark Integration

Gauntlet is included in Terrain's benchmark matrix as a real-world repository:

```json
{
  "name": "gauntlet",
  "path": "benchmarks/repos/gauntlet",
  "type": "real",
  "description": "AI eval execution provider — deterministic eval runner with baseline comparison"
}
```

This ensures Terrain's analysis engines (framework detection, signal generation, coverage analysis) are validated against Gauntlet's codebase structure.

## Future: `terrain ai run` Integration

When `terrain ai run` is fully implemented, it will:

1. Select relevant scenarios based on code change (`terrain impact`)
2. Invoke Gauntlet as the execution backend
3. Ingest results automatically
4. Report pass/fail with explanation traces

The interface between Terrain and Gauntlet will be the artifact JSON described above — Terrain writes a scenario selection file, Gauntlet executes and writes results, Terrain ingests results.
