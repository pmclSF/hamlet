# Testing and Quality: Contributor Guide

This guide covers how to write tests for Terrain's Go codebase, when to use
each test type, and how to work with fixtures and golden files.

## Decision Tree: Which Test to Write

```
Is this a pure function or method with clear inputs/outputs?
  --> Table-driven unit test

Does it integrate multiple subsystems (detector + measurement + render)?
  --> Fixture-based integration test

Does it produce formatted output that should not change unexpectedly?
  --> Golden test (assertGolden)

Must the output be identical across repeated runs?
  --> Determinism test

Does it handle malformed, empty, or adversarial input?
  --> Adversarial test
```

When in doubt, start with a table-driven unit test. Add integration or
golden tests when the unit under test has complex output or crosses
subsystem boundaries.

## Common Patterns

### Table-Driven Unit Tests

The standard pattern for testing functions with multiple input/output cases:

```go
func TestDetectFramework(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"go test file", "func TestFoo(t *testing.T)", "go"},
        {"jest test file", "describe('foo', () => {", "jest"},
        {"no framework", "package main", ""},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := DetectFramework(tt.input)
            if got != tt.expected {
                t.Errorf("got %q, want %q", got, tt.expected)
            }
        })
    }
}
```

### Fixture-Based Integration Tests

Load a fixture, run a pipeline stage, and assert on the result:

```go
func TestAnalyzeSampleRepo(t *testing.T) {
    snapshot := loadFixture(t, "sample-repo")
    result := analysis.Analyze(snapshot)
    if len(result.Signals) == 0 {
        t.Fatal("expected signals, got none")
    }
}
```

### The assertGolden Helper

Golden tests compare output against a checked-in reference file:

```go
func TestRenderSummary(t *testing.T) {
    snapshot := loadFixture(t, "sample-repo")
    output := report.RenderSummaryReport(snapshot)
    assertGolden(t, "summary-report", output)
}
```

On first run or when updating, pass `-update` to write the golden file.
The golden file is stored alongside the test under `testdata/golden/`.

## How-To Guides

### Adding a New Fixture

1. Create a directory under `internal/testdata/` with representative files.
2. Include the minimum set of files needed to exercise the target behavior.
3. Add a `README` or comment explaining what the fixture tests.
4. Reference the fixture from your test using the standard load helper.

### Adding a Golden Test

1. Write the test using `assertGolden(t, "descriptive-name", output)`.
2. Run with `-update` to generate the initial golden file.
3. Review the generated file carefully before committing.
4. On future runs, any output change will fail the test until the golden
   file is explicitly updated.

### Adding a Determinism Test

1. Run the operation twice in the same test with identical input.
2. Compare outputs byte-for-byte.
3. If the operation involves maps or goroutines, run it 10+ times to
   surface non-deterministic ordering.

### Adding an Adversarial Test

1. Identify an input boundary: empty file, binary content, extremely
   long lines, deeply nested structures, circular references.
2. Write a test that passes this input and asserts the function either
   returns a meaningful result or a clear error (never panics).
3. Place adversarial tests in the same test file as the unit they target.

### Running Benchmarks

```bash
go test ./... -bench=. -benchmem           # All benchmarks
go test ./internal/analysis/ -bench=BenchmarkAnalyze  # Specific benchmark
go test ./... -bench=. -count=5            # Multiple iterations for stability
```

### Updating Golden Files

```bash
go test ./... -run TestGolden -update      # Update all golden files
go test ./internal/report/ -run TestRenderSummary -update  # Update one
```

Always review the diff after updating golden files. Unexpected changes
indicate a regression.

## Subsystem-to-Test-Layer Mapping

| Subsystem | Unit Tests | Integration | Golden | E2E |
|-----------|-----------|-------------|--------|-----|
| Detectors | Per-detector logic | Detector + signals | Signal output format | Full pipeline |
| Measurements | Scoring functions | Measurement + posture | Posture report | Full pipeline |
| Renderers | Formatting helpers | Render + data | Report output | Full pipeline |
| CLI | Flag parsing | Command execution | Help text | CLI regression |
| Impact | Graph traversal | Impact + ownership | Impact report | Full pipeline |
| Export | Field filtering | Export + privacy | JSON schema | E2E privacy |
| Ownership | CODEOWNERS parsing | Ownership + aggregation | Owner report | Full pipeline |

## Checklist Before Submitting

- [ ] `go vet ./...` passes
- [ ] `go test ./...` passes
- [ ] New public functions have test coverage
- [ ] Golden files are reviewed if updated
- [ ] No test depends on external network or filesystem state
- [ ] Adversarial cases are covered for user-facing input boundaries
