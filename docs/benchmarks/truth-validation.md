# Truth Validation

> **Purpose:** Validate Terrain's analysis precision and recall against documented ground truth.
> **Harness:** `cmd/terrain-truthcheck/`
> **Library:** `internal/truthcheck/`

## Overview

The truth validation harness compares Terrain's actual analysis output against a ground truth specification. It runs the full pipeline, then evaluates each analysis category with precision/recall scoring.

## Usage

```bash
# Run against terrain-world fixture
go run ./cmd/terrain-truthcheck \
  --root tests/fixtures/terrain-world \
  --truth tests/fixtures/terrain-world/tests/truth/terrain_truth.yaml \
  --output benchmarks/output/truthcheck/

# JSON output only
go run ./cmd/terrain-truthcheck \
  --root tests/fixtures/terrain-world \
  --truth tests/fixtures/terrain-world/tests/truth/terrain_truth.yaml \
  --json
```

## Truth Spec Format

The truth spec is a YAML file documenting expected findings per category:

```yaml
coverage:
  description: "Coverage analysis should identify structural gaps"
  expected_uncovered:
    - path: "src/payments/subscription.ts"
      reason: "No unit tests exist"

redundancy:
  expected_clusters:
    - tests: ["tests/unit/a.test.ts", "tests/unit/b.test.ts"]
      reason: "Near-identical test files"

fanout:
  expected_flagged:
    - node: "src/shared-db.ts"
      expected_min_dependents: 6

ai:
  expected_scenarios: 4
  expected_prompt_surfaces: [...]
  expected_dataset_surfaces: [...]
```

## Evaluation Categories

| Category | What It Validates | Scoring Method |
|----------|------------------|----------------|
| **impact** | Changed file → impacted tests/scenarios | Per-test recall: expected tests found in impact result |
| **coverage** | Uncovered/weakly-covered files | Set matching: expected paths vs actual weak areas |
| **redundancy** | Duplicate test clusters | Cluster containment: expected test pairs found in actual clusters |
| **fanout** | High-fanout nodes | Node matching: expected nodes found in flagged list |
| **stability** | Skip/flaky patterns | Signal matching: expected files in skip/flaky signals |
| **ai** | Scenarios, prompts, datasets | Count + set matching: expected counts and surface IDs |
| **environment** | Platform/environment coverage | Informational (no hard assertions) |

## Scoring

Each category produces:

| Metric | Formula | Description |
|--------|---------|-------------|
| **Precision** | `matched / (matched + unexpected)` | How many found items were expected |
| **Recall** | `matched / (matched + missing)` | How many expected items were found |
| **F1 Score** | `2 * P * R / (P + R)` | Harmonic mean of precision and recall |
| **Pass** | `recall >= 0.5` | Category passes if at least half of expectations met |

The overall report averages scores across all categories.

## Output

### `report.json`

```json
{
  "repoRoot": "/path/to/repo",
  "truthFile": "/path/to/truth.yaml",
  "categories": [
    {
      "category": "coverage",
      "passed": true,
      "score": 1.0,
      "precision": 1.0,
      "recall": 1.0,
      "expected": 2,
      "found": 2,
      "matched": 2,
      "missing": [],
      "unexpected": []
    }
  ],
  "summary": {
    "totalCategories": 7,
    "passedCount": 6,
    "overallScore": 0.93,
    "overallPrecision": 0.86,
    "overallRecall": 0.86
  }
}
```

### `report.md`

Markdown report with summary table, per-category results, missing items, and unexpected items.

## Current Results (terrain-world fixture)

| Category | Score | Precision | Recall | Status |
|----------|-------|-----------|--------|--------|
| coverage | 100% | 100% | 100% | PASS |
| redundancy | 100% | 100% | 100% | PASS |
| fanout | 100% | 100% | 100% | PASS |
| stability | 50% | — | — | FAIL (requires runtime data) |
| ai | 100% | 100% | 100% | PASS |
| impact | 100% | 100% | 100% | PASS |
| environment | 100% | 100% | 100% | PASS |
| **Overall** | **93%** | **86%** | **86%** | **6/7 passed** |

## Key Types

```go
// TruthCategoryResult is the validation result for one truth category.
type TruthCategoryResult struct {
    Category    string   // e.g., "coverage", "impact", "ai"
    Passed      bool     // recall >= 0.5
    Score       float64  // F1 score (0.0-1.0)
    Precision   float64  // correct / (correct + unexpected)
    Recall      float64  // correct / (correct + missing)
    Expected    int      // total expected items
    Found       int      // total found items
    Matched     int      // correctly matched
    Missing     []string // expected but not found
    Unexpected  []string // found but not expected
}

// TruthCheckReport is the complete validation result.
type TruthCheckReport struct {
    RepoRoot   string
    TruthFile  string
    Categories []TruthCategoryResult
    Summary    ReportSummary
}
```

## Adding New Truth Specs

1. Create a `tests/truth/terrain_truth.yaml` in your fixture
2. Document expected findings per category
3. Run `terrain-truthcheck --root <fixture> --truth <truth.yaml>`
4. Review missing/unexpected items and adjust truth spec or fix detection
