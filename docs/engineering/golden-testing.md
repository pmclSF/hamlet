# Golden Testing

Golden tests capture the full output of a subsystem as a reference file, then
compare future runs against that baseline. They detect unintended regressions in
output format, content, and structure without writing brittle per-field assertions.

## Infrastructure

All golden test infrastructure lives in `internal/testdata/golden_test.go`.

### The `assertGolden` Helper

```go
func assertGolden(t *testing.T, name string, got []byte) {
    path := goldenPath(name)  // golden/<name>
    if *update {
        os.WriteFile(path, got, 0o644)
        return
    }
    want, _ := os.ReadFile(path)
    if !bytes.Equal(got, want) {
        t.Errorf("golden mismatch for %s\n--- want (first 500 chars) ---\n%s\n--- got (first 500 chars) ---\n%s",
            name, truncate(want, 500), truncate(got, 500))
    }
}
```

Key behaviors:
- **`-update` flag.** Pass `-update` to overwrite golden files with current output.
- **Truncated diffs.** Mismatches show the first 500 characters of both expected
  and actual output, keeping test failure messages readable.
- **Byte-exact comparison.** Uses `bytes.Equal`, so whitespace, newlines, and
  field ordering all matter.

### Golden Directory

Golden files are stored in `internal/testdata/golden/` and committed to the
repository. They are part of the source of truth for expected output.

## Golden Test Inventory

| Test Function | Golden File | Fixture Used | Subsystems Covered |
|---|---|---|---|
| `TestGolden_MetricsJSON` | `metrics-minimal.json` | MinimalSnapshot | metrics.Derive, JSON serialization |
| `TestGolden_ExportJSON` | `export-minimal.json` | MinimalSnapshot | benchmark.BuildExport, measurements, portfolio |
| `TestGolden_AnalyzeText` | `analyze-minimal.txt` | MinimalSnapshot | reporting.RenderAnalyzeReport, risk, measurements, portfolio |
| `TestGolden_PortfolioText` | `portfolio-flaky.txt` | FlakyConcentratedSnapshot | reporting.RenderPortfolioReport, risk, measurements, portfolio |
| `TestGolden_SummaryText` | `summary-minimal.txt` | MinimalSnapshot | reporting.RenderSummaryReport, heatmap, risk, measurements |
| `TestGolden_PostureText` | `posture-minimal.txt` | MinimalSnapshot | reporting.RenderPostureReport, measurements |

### Planned Golden Tests

| Test Function | Golden File | Purpose |
|---|---|---|
| `TestGolden_ImpactJSON` | `impact-healthy.json` | Capture impact analysis output for a standard change scope |
| `TestGolden_CompareText` | `compare-migration.txt` | Capture comparison report between two snapshots over time |

## Determinism Requirements

Golden tests depend on fully deterministic output. All time-dependent fields are
normalized to `FixedTime` (2025-01-15T12:00:00Z) before serialization. Any
subsystem that introduces non-determinism (map iteration order, goroutine
scheduling, wall-clock timestamps) will cause golden test failures.

## Workflow: Reviewing and Updating Golden Files

1. **Run tests.** `go test ./internal/testdata/ -run TestGolden`
2. **See the diff.** If a golden test fails, the truncated diff shows what changed.
   For the full diff, compare the golden file against the new output manually.
3. **Review intent.** Determine whether the change is intentional (new feature,
   improved formatting) or a regression. This is the critical step -- golden
   files should only change when you intend them to.
4. **Update with `-update`.** If the change is intentional:
   ```bash
   go test ./internal/testdata/ -run TestGolden -update
   ```
5. **Inspect the updated files.** Review the diff in `git diff` before committing.
   Never blindly accept golden file updates.
6. **Commit golden files.** Golden files are source-controlled artifacts. Include
   them in the same commit as the code change that caused the update.

## When to Add a New Golden Test

Add a golden test when:
- A new report renderer is added (text or JSON output)
- A new export format is introduced
- A CLI command produces structured output that users depend on

Do not add golden tests for:
- Internal data structures that are not user-facing
- Intermediate computation results
- Subsystems already covered by determinism tests

## Relationship to Other Test Categories

Golden tests complement but do not replace:
- **Determinism tests** verify N-run identity without a stored reference file.
- **Schema tests** verify round-trip serialization fidelity.
- **Adversarial tests** verify graceful degradation, not output correctness.

Golden tests are the only category that stores expected output on disk and requires
explicit human review when output changes.
