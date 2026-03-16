# python-ml-observatory

Python ML platform fixture with 6 domains, pytest eval suites, and 7 intentional problems.

## Domains

| Domain | Source | Purpose |
|--------|--------|---------|
| data | `src/data/` | Loaders, transforms, augmentation |
| models | `src/models/` | Classifier, embeddings |
| retrieval | `src/retrieval/` | Search, re-ranking |
| prompts | `src/prompts/` | Prompt building |
| scoring | `src/scoring/` | Metrics, batch evaluation |
| safety | `src/safety/` | Input/output safety filters |

## Intentional Problems

1. **Duplicate eval files** — `test_classifier_accuracy.py` and `test_classifier_eval_v2.py`
2. **Duplicate safety evals** — `test_prompt_safety.py` and `test_safety_regression.py`
3. **Untested augmentation** — `src/data/augment.py` has no tests
4. **Untested batch scoring** — `src/scoring/batch.py` has no direct test imports
5. **Fanout via shared helpers** — `conftest_helpers.py` imported by multiple test files
6. **Overlapping AI scenarios** — classifier-accuracy and classifier-eval-v2 cover same surfaces
7. **Overlapping safety scenarios** — prompt-safety and safety-regression cover same surfaces

## Omitted Truth Categories

- **stability** — no runtime artifacts; skip/flaky detection requires `--runtime` data
