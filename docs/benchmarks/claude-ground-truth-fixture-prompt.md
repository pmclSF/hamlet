# Claude Ground-Truth Fixture Prompt

Use this from the repository root when you want Claude Code to generate new
fixture repositories that match Terrain's existing benchmark structure and
validate themselves with the truth-check evaluator.

As of March 15, 2026, the existing `terrain-world` fixture still exits non-zero
under `go run ./cmd/terrain-truthcheck ...` because its `stability` category is
not fully green without supporting runtime artifacts. The prompt below is
written to avoid that trap: it tells Claude to keep only truth categories that
can be made fully green for each new fixture.

## Prompt

```text
You are working inside the Terrain repository at its root.

Goal: create 3 new, highly complex fixture repositories under `tests/fixtures/`
with precise ground-truth specs, register them in `benchmarks/repos.json`, and
use the existing truth evaluator to iterate until every new fixture passes.

First, read and follow the existing patterns in:
- `tests/fixtures/terrain-world/`
- `tests/fixtures/ai-eval-suite/`
- `tests/fixtures/high-fanout/`
- `tests/fixtures/duplicate-heavy/`
- `tests/fixtures/qa-manual-overlay/`
- `docs/benchmarks/truth-validation.md`
- `benchmarks/repos.json`

Non-negotiable structure for each new fixture:
- Directory: `tests/fixtures/<fixture-name>/`
- `README.md` describing domains, validation layers, intentional problems, and
  truth coverage
- Realistic source tree with enough imports and shared helpers to create
  meaningful impact chains
- Tests across multiple layers where appropriate: unit, integration, e2e,
  contract, eval, manual coverage, etc.
- `.terrain/terrain.yaml` with declared `scenarios` and `manual_coverage` when
  relevant
- `tests/truth/terrain_truth.yaml` with explicit expectations
- Any supporting files the repo shape needs: `package.json`, `pyproject.toml`,
  `go.mod`, `CODEOWNERS`, coverage artifacts, runtime artifacts, and similar
- Add an entry for each fixture to `benchmarks/repos.json` with `type:
  "fixture"` and an honest description

Create these 3 fixtures:

1. `saas-control-plane`
- TypeScript-first B2B SaaS platform with auth, billing, entitlements, audit,
  search, notifications, and ai-assistant domains
- Mix unit, integration, e2e, eval, and manual coverage
- Intentional issues: shared database fanout, duplicate purchase and admin
  tests, weak coverage in billing and notifications, overlapping AI safety
  scenarios, one manual-only critical area, one policy-sensitive ownership
  boundary

2. `python-ml-observatory`
- Python-heavy ML platform with data loaders, classifier, embeddings,
  retrieval, prompt builder, batch scoring, and safety filters
- Pytest eval suites plus datasets and scenario declarations
- Intentional issues: duplicated eval files, partially untested dataset
  utilities, fanout via a shared fixture module, and one quarantined or skipped
  area only if you can make truthcheck green with supporting artifacts;
  otherwise do not include a stability truth section

3. `legacy-omnichannel`
- Mixed JS/TS commerce repo with Jest, Mocha, and Cypress-era legacy overlap,
  plus a small AI merchandising module
- Domains: cart, checkout, fraud, refunds, mobile, recommendations, admin
- Intentional issues: migration blockers, duplicate checkout flows,
  high-fanout shared helpers, uncovered mobile and refund paths, overlapping
  scenario coverage on prompt and dataset surfaces

Complexity bar for every fixture:
- At least 15 source files and 12 validation files
- At least 5 business domains
- At least 6 intentional, documented problems
- Enough import and dependency structure that `terrain impact` can find
  multi-hop results
- Ground truth must use exact paths and scenario names, not vague descriptions

Truth spec guidance:
- Prefer categories that can be made fully green with the current Terrain
  engine: `impact`, `coverage`, `redundancy`, `fanout`, `ai`, and
  `environment`
- Include `stability` only when the fixture has the supporting runtime or code
  patterns needed for Terrain to pass it reliably
- If a category is omitted, explain why in the fixture `README.md`

Evaluator rules:
- Use the existing evaluator only:
  `go run ./cmd/terrain-truthcheck --root <fixture> --truth <fixture>/tests/truth/terrain_truth.yaml --json`
- The evaluator exits non-zero when any category fails. Treat that as a bug to
  fix in the fixture, its truth spec, or its supporting artifacts.
- Do not weaken expectations just to get green. Prefer fixing the repository
  shape so Terrain can discover the intended behavior.
- Do not modify `cmd/terrain-truthcheck` or `internal/truthcheck` unless you
  find a real evaluator bug. If you believe there is an evaluator bug, stop and
  document it instead of silently editing the evaluator.

Execution order:
1. Build one fixture completely before starting the next.
2. After creating a fixture, run the evaluator for that fixture.
3. If the evaluator fails, inspect the missing and unexpected items from the
   JSON report and fix the fixture.
4. Repeat until that fixture passes.
5. Move to the next fixture and repeat.
6. After all 3 pass individually, run a final sweep across all new fixtures and
   fix any regressions.

Suggested verification loop:
- `go run ./cmd/terrain-truthcheck --root tests/fixtures/saas-control-plane --truth tests/fixtures/saas-control-plane/tests/truth/terrain_truth.yaml --json`
- `go run ./cmd/terrain-truthcheck --root tests/fixtures/python-ml-observatory --truth tests/fixtures/python-ml-observatory/tests/truth/terrain_truth.yaml --json`
- `go run ./cmd/terrain-truthcheck --root tests/fixtures/legacy-omnichannel --truth tests/fixtures/legacy-omnichannel/tests/truth/terrain_truth.yaml --json`

Working style:
- Follow the repo's existing naming, doc style, and fixture conventions.
- Prefer ASCII.
- Do not create placeholder files with toy logic or lorem ipsum behavior.
- Keep the fixture repos deterministic and self-contained.
- Document the intentional problems clearly in each `README.md` so the truth
  spec is auditable.
- Update `benchmarks/repos.json` as you go, not only at the end.

Deliverables before you stop:
- The 3 new fixture repos
- Updated `benchmarks/repos.json`
- A short markdown summary at
  `docs/benchmarks/generated-ground-truth-fixtures.md` listing each new
  fixture, its themes, and the final evaluator status
- Final command results summarized with pass/fail status per fixture

Do not stop after scaffolding. Finish the repos, run the evaluator, and loop
until every new fixture passes.
```
