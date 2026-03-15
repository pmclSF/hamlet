# Terrain Product Positioning

## What Terrain is

Terrain is a signal-first test intelligence tool that analyzes your test suite and tells you what actually matters — which areas are risky, which tests are weak, and where to invest next.

## Who it is for

- **Engineering teams** maintaining large or growing test suites
- **Tech leads** planning test modernization or framework migration
- **Engineering managers** who need to understand test health without reading every file
- **Platform teams** responsible for test infrastructure quality

## What problems it solves

1. **Invisible test quality decay.** Tests accumulate weak assertions, heavy mocking, dead code, and framework fragmentation. Terrain makes these patterns visible.

2. **Migration risk blindness.** Teams attempt framework migrations without understanding which files have blockers, which areas compound migration risk with quality issues, and where to start safely.

3. **Flakiness without focus.** Teams know tests are flaky but treat it as a systemic problem. Terrain shows that flakiness often concentrates in a handful of files.

4. **Coverage theater.** High line-coverage numbers mask the fact that exported functions have no unit tests, assertions are weak, and coverage diversity is poor.

5. **No structural understanding.** Traditional tools count lines and branches. Terrain reasons about code units, frameworks, ownership, and risk dimensions.

## How it differs from coverage tools

| Coverage tools | Terrain |
|---|---|
| Line/branch counts | Structural risk and posture dimensions |
| Pass/fail reporting | Signal-first analysis with evidence and confidence |
| One number | Five posture dimensions with explainability |
| Per-file view | Directory, owner, and repo-level risk surfaces |
| No migration awareness | Migration readiness with area assessments |
| No quality analysis | Assertion density, mock concentration, dead tests |

## Design philosophy

**Inference-first.** Terrain reads your repo and infers structure from what already exists — import graphs, file naming, coverage artifacts, runtime results. No annotations, no tagging, no SDK integration required. Run `terrain analyze` and get a complete assessment.

**Explainability-first.** Every finding carries an evidence chain. `terrain explain` traces any decision back to the signals, dependency paths, and scoring rules that produced it. No black boxes, no magic numbers, no opaque scores.

## What Terrain does that no one else does

Terrain treats the test suite as a portfolio of investments with measurable cost, protection value, and efficiency. The `terrain portfolio` command surfaces which tests provide the most leverage, which waste CI resources through redundant coverage, and where runtime concentrates — turning test maintenance from guesswork into portfolio management.

## One sentence

Terrain looks at your test suite the way a senior engineer would — finding the structural issues that line counts miss.
