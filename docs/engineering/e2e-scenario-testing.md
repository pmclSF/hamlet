# E2E Scenario Testing

Terrain's end-to-end tests exercise full pipeline flows from analysis through
rendering, validating that the system produces correct outputs for real-world
workflows. These tests live in `internal/testdata/e2e_test.go`.

## Fixture-Driven Approach

E2E tests use fixture snapshots checked into the repository under
`internal/testdata/`. Each fixture represents a known repository state with
pre-computed signals, coverage data, and ownership mappings. Tests load a
fixture, run it through the full pipeline, and assert on the output.

This approach provides several guarantees:

- **Reproducibility.** Fixtures are static, so tests always run against the
  same input regardless of environment.
- **Speed.** No git cloning, no network calls, no file-system scanning.
  Fixtures load in microseconds.
- **Determinism.** Combined with `models.SortSnapshot()`, fixture-based tests
  produce byte-identical output across runs.

## Flagship Workflows

The E2E suite covers the following primary workflows:

| Workflow | What It Validates |
|----------|-------------------|
| Analyze + Summary + Render | Full pipeline from raw analysis through summary generation to human-readable report rendering |
| Comparison Workflow | Two-snapshot comparison producing meaningful diffs of signals, risk, and coverage changes |
| Impact Workflow | Change-scoped impact analysis identifying affected test files and protection gaps |
| Migration Risk Flow | Migration readiness scoring with quality band assignment and signal aggregation |
| Graph-Enriched Heatmap | Risk heatmap generation using dependency graph data for structural context |
| Ownership Propagation + Aggregation | CODEOWNERS parsing, ownership assignment to test files, and roll-up to team-level summaries |
| Portfolio Intelligence | Multi-repository aggregation producing cross-project posture comparisons |
| Export Privacy | Ensures that exported data (JSON, benchmark matrices) excludes sensitive fields and respects privacy boundaries |

## Validation Strategy

E2E tests validate two categories of output:

### Machine-Readable Outputs

- **JSON roundtrip.** Serialize the snapshot to JSON, deserialize it back, and
  compare field-by-field to confirm no data loss.
- **Counts.** Assert exact counts for test files, test cases, signals, risk
  items, and coverage entries. These catch regressions where a detector
  silently stops emitting findings.
- **Schema compliance.** Output conforms to the schemas defined in
  `docs/schema/` so downstream consumers (VS Code extension, CI reporters)
  can rely on stable structure.

### User-Facing Summaries

- **String contains checks.** Rendered reports must include expected section
  headers, metric labels, and key data points (e.g., posture band name,
  signal counts, risk categories).
- **Absence checks.** Certain internal fields, debug output, or sensitive
  data must not appear in user-facing renders.
- **Formatting stability.** Table alignment, column headers, and separator
  lines must match expected patterns.

## Adding a New E2E Scenario

1. Create or update a fixture under `internal/testdata/` with the input data
   your scenario requires.
2. Add a test function in `e2e_test.go` following the existing pattern:
   load fixture, run pipeline, assert on output.
3. Include both machine-readable assertions (counts, JSON structure) and
   user-facing assertions (string contains) for complete coverage.
4. Run `go test ./internal/testdata/ -run TestE2E` to verify.

## Relationship to Other Test Layers

E2E tests sit at the top of the test pyramid. They depend on unit-tested
detectors, measurements, and renderers functioning correctly. When an E2E
test fails, the root cause is usually in a lower layer -- use the E2E
failure message to identify which subsystem to investigate, then write a
targeted unit test to reproduce and fix the bug.

See also:
- `docs/engineering/verification-system-map.md` for the full test layer diagram
- `docs/engineering/determinism.md` for the deterministic output contract
- `docs/engineering/cli-and-docs-regression-testing.md` for CLI-level tests
