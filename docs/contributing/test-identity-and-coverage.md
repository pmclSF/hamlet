# Extending Test Identity and Coverage

This guide covers how to extend Hamlet's test identity, type inference, coverage ingestion, and attribution systems.

## Adding a New Language for Test Extraction

### 1. Add extraction in `internal/testcase/extract.go`

Create an `extractLang()` function following the existing pattern:

```go
func extractRuby(src, relPath, framework string) []TestCase {
    lines := strings.Split(src, "\n")
    var cases []TestCase
    // Parse test definitions, build TestCase structs.
    // Set: TestName, SuiteHierarchy, Line, ExtractionKind, Confidence
    // Do NOT set: TestID, CanonicalIdentity, FilePath, Framework, Language
    // (these are assigned automatically by Extract())
    return cases
}
```

### 2. Register in the language switch

In `Extract()`, add a case to the language switch:

```go
case "ruby":
    cases = extractRuby(src, relPath, framework)
```

### 3. Map framework to language

In `FrameworkLanguage()`:

```go
case "rspec", "minitest":
    return "ruby"
```

### 4. Add tests

Create test cases in `internal/testcase/extract_test.go` covering:
- Basic extraction with suite hierarchy
- Stable IDs across runs
- Line movement without rename → same ID
- Parameterized test handling if applicable

## Adding Test Type Inference Rules

### 1. Add a rule function in `internal/testtype/infer.go`

```go
func inferFromImports(filePath string, content string) InferResult {
    // Return InferResult with Type, Confidence, Evidence
}
```

### 2. Register in `InferForTestCase()`

Add your rule call in the candidates list:

```go
if r := inferFromImports(tc.FilePath, content); r.Type != TypeUnknown {
    candidates = append(candidates, r)
}
```

### 3. Confidence guidelines

| Signal strength | Confidence |
|----------------|------------|
| Framework/tooling match | 0.8–0.9 |
| Path convention match | 0.7–0.85 |
| Name/annotation match | 0.6–0.7 |
| Weak heuristic | 0.4–0.5 |

## Adding a New Coverage Format

### 1. Add a parser in `internal/coverage/ingest.go`

```go
func parseClover(data []byte) ([]CoverageRecord, error) {
    // Parse the format into []CoverageRecord
    // Populate LineHits, BranchHits, FunctionHits as available
    // Call recomputeCounts() on each record
}
```

### 2. Add format detection

In `IngestFile()`, add a detection branch:

```go
if strings.Contains(trimmed, "<coverage") {
    records, err = parseClover(data)
    format = "clover"
}
```

### 3. Add to `isCoverageFile()`

```go
strings.HasSuffix(name, ".xml") || name == "clover.xml"
```

### 4. Test with fixtures

Add test data and tests in `internal/coverage/ingest_test.go`.

## Extending Coverage Attribution

### Adding per-unit branch coverage

Currently branch coverage is file-level. To make it per-unit:

1. In `AttributeToCodeUnits()`, scope branch hits to the unit's line range
2. Update `BranchCoveragePct` computation to use only in-range branches
3. Update `EvidenceQuality` to "exact" when branch data is scoped

### Adding per-test coverage ingestion

1. Parse per-test coverage artifacts into `TestCoverageRecord` structs
2. Call `BuildPerTestCoverage()` with the records and a `UnitSpan` index
3. Attach `PerTestCoverage` results to the snapshot

## Key Invariants

When extending these systems, maintain these invariants:

| Invariant | Where enforced |
|-----------|---------------|
| Test IDs are deterministic | `identity.GenerateID()` uses SHA-256 |
| Identity is structural, not positional | Never hash line numbers |
| Collisions are detected | `testcase.DetectAndResolveCollisions()` |
| Evidence quality is explicit | `UnitCoverage.EvidenceQuality` |
| Snapshot output is deterministic | `models.SortSnapshot()` |
| Coverage degrades gracefully | `-1` for unavailable, never fake zeros |
| Type inference is explainable | `InferResult.Evidence` list |

## Package Reference

| Package | Purpose |
|---------|---------|
| `internal/identity` | Path/name normalization, canonical identity, hash generation |
| `internal/testcase` | TestCase model, extraction, collision detection |
| `internal/testtype` | Test type inference with evidence |
| `internal/coverage` | Coverage ingestion, attribution, by-type, per-test, insights |
| `internal/models` | Snapshot-level models (TestCase, CodeUnit, CoverageSummary) |
| `internal/comparison` | Snapshot comparison including TestCaseDeltas and CoverageDelta |
