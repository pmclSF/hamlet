# Example: `terrain analyze` — AI Evaluation Suite (Python)

> Scenario: A Python ML project with pytest-based evaluation tests, dataset loaders, and model inference functions. No scenarios declared yet, no coverage or runtime artifacts.

## Terminal Output

```
Terrain — Test Suite Analysis
============================================================

Repository Profile
------------------------------------------------------------
  Test volume:          small
  CI pressure:          low
  Coverage confidence:  high
  Redundancy level:     low
  Fanout burden:        low

Validation Inventory
------------------------------------------------------------
  Test files:     6
  Test cases:     18
  Code units:     9
  Code surfaces:  9
  Datasets:       2
  Frameworks:
    pytest                  4 files [unit]

Coverage Confidence
------------------------------------------------------------
  High:    4 files (100%)
  Medium:  0 files (0%)
  Low:     0 files (0%)

High-Fanout Nodes
------------------------------------------------------------
  Flagged: 2 (threshold: 10)
      (behavior_surface, 12 dependents)
      (behavior_surface, 12 dependents)

CI Optimization Potential
------------------------------------------------------------
  High-fanout nodes:          2
  2 high-fanout nodes could be refactored to reduce blast radius.

Top Insight
------------------------------------------------------------
   fans out to 12 transitive dependents — changes here trigger wide test impact. Consider splitting or isolating.

Risk Posture
------------------------------------------------------------
  health:                  STRONG
  coverage_depth:          STRONG
  coverage_diversity:      STRONG
  structural_risk:         STRONG
  operational_risk:        STRONG

Signals: 0 total

Behavior Redundancy
------------------------------------------------------------
  Redundant tests:  18 across 3 clusters
  [wasteful] 8 tests, 4 shared surfaces (100% confidence) [pytest]
         Tests in the same framework exercise 4 identical behavior surfaces (100% overlap). Consolidation would reduce CI cost without losing coverage.
  [wasteful] 5 tests, 12 shared surfaces (100% confidence) [pytest]
         Tests in the same framework exercise 12 identical behavior surfaces (100% overlap). Consolidation would reduce CI cost without losing coverage.
  [wasteful] 5 tests, 4 shared surfaces (100% confidence) [pytest]
         Tests in the same framework exercise 4 identical behavior surfaces (100% overlap). Consolidation would reduce CI cost without losing coverage.

Stability
------------------------------------------------------------
  No runtime data provided. Provide --runtime (JUnit XML or Jest JSON)
  to unlock: flaky test detection, slow test flagging, stability clustering.

Edge Cases
------------------------------------------------------------
  [warning] CI is already fast — optimization may yield minimal benefit.

Policy Recommendations
------------------------------------------------------------
  • CI is already fast. Test selection would yield minimal time savings.

Data Completeness
------------------------------------------------------------
  [available] source
  [missing  ] coverage
  [missing  ] runtime
  [missing  ] policy

Limitations
------------------------------------------------------------
  * No coverage data provided; coverage confidence is structural (import-based) only.
  * No policy file found; governance checks skipped.
  * No runtime data provided; skip/flaky/slow test detection unavailable.

Next steps:
  terrain analyze --verbose           show all findings
  terrain impact                      what tests matter for this change?
  terrain explain <test-path>         why was a test selected?
  terrain insights                    deeper analysis with recommendations
```

## What Each Section Shows for AI Repos

| Section | What It Answers | Key Insight in This Example |
|---------|----------------|----------------------------|
| **Validation Inventory** | What validations exist? | 18 test cases, 9 code surfaces, **2 dataset surfaces** detected |
| **Coverage Confidence** | How well are source files linked to tests? | 100% high — all eval files structurally linked |
| **Behavior Redundancy** | Are eval tests overlapping? | 18 tests in 3 clusters — significant overlap in eval coverage |
| **High-Fanout Nodes** | What's fragile? | 2 behavior surfaces with 12 dependents — shared eval infrastructure |
| **Stability** | Are evals reliable? | No runtime data yet — hint to provide artifacts |

## AI-Specific Detection

Terrain automatically infers AI-relevant code surfaces:

### Datasets

Functions and variables matching dataset patterns are classified as `SurfaceDataset`:

```python
# Detected as dataset surfaces:
def load_dataset(path):        # name contains "dataset"
def training_data():           # name contains "training_data"
eval_data = load_csv("...")    # name contains "eval_data"
```

### Prompts

Functions and variables matching prompt patterns are classified as `SurfacePrompt`:

```python
# Detected as prompt surfaces:
def build_prompt(context):     # name contains "prompt"
system_template = "You are..." # name contains "template"
```

When prompts or datasets appear in the validation inventory, they participate in:
- **Impact analysis:** Changing a prompt file surfaces impacted scenarios via `terrain impact`
- **Behavior grouping:** Related prompts are grouped into behavior surfaces
- **Coverage analysis:** Untested prompts/datasets appear in weak coverage areas

## Next Steps for AI Repos

```bash
# See what AI-specific assets Terrain detected
terrain ai list

# Validate AI/eval setup
terrain ai doctor

# Add scenarios for richer coverage analysis
# (in .terrain/terrain.yaml)

# Check impact when a prompt changes
terrain impact --base main
```

## JSON Output (`--json`)

```json
{
  "testsDetected": {
    "testFileCount": 6,
    "testCaseCount": 18,
    "codeUnitCount": 9,
    "codeSurfaceCount": 9,
    "datasetCount": 2,
    "frameworks": [
      { "name": "pytest", "fileCount": 4, "type": "unit" }
    ]
  },
  "coverageConfidence": {
    "highCount": 4,
    "mediumCount": 0,
    "lowCount": 0,
    "totalFiles": 4
  },
  "highFanout": {
    "flaggedCount": 2,
    "threshold": 10
  },
  "behaviorRedundancy": {
    "redundantTestCount": 18,
    "clusters": [
      {
        "tests": ["test:evals/accuracy/test_accuracy.py:..."],
        "sharedSurfaces": ["surface:src/model.py:predict", "..."],
        "confidence": 1.0,
        "overlapKind": "wasteful",
        "frameworks": ["pytest"]
      }
    ]
  },
  "riskPosture": [
    { "dimension": "health", "band": "strong" },
    { "dimension": "coverage_depth", "band": "strong" },
    { "dimension": "coverage_diversity", "band": "strong" },
    { "dimension": "structural_risk", "band": "strong" },
    { "dimension": "operational_risk", "band": "strong" }
  ]
}
```
