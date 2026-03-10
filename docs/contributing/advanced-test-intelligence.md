# Contributing: Advanced Test Intelligence Subsystems

This guide covers how to extend the advanced assessment and workflow subsystems in Hamlet's current engine. It describes how to add new assessment dimensions, new CLI drill-down entities, new output formats, and how to follow the conventions that keep these subsystems consistent and cautious.

## Package Conventions

Each assessment subsystem follows a consistent file layout:

```
internal/<subsystem>/
  model.go              # Types: enums (typed string constants), result structs, input structs
  <operation>.go        # Core logic: detect.go, classify.go, assess.go, analyze.go, or infer.go
  <subsystem>_test.go   # Tests covering happy path, edge cases, and empty/nil inputs
```

Additional files appear when the subsystem has supporting logic:

- `similarity.go` (lifecycle) -- string/path similarity functions
- `build.go` (stability) -- history construction from snapshot sequences
- `render.go` (changescope) -- output rendering functions

The naming convention for the logic file reflects the primary verb of the subsystem:

| Subsystem | Logic File | Primary Function |
|-----------|-----------|-----------------|
| lifecycle | `inference.go` | `InferContinuity(from, to *models.TestSuiteSnapshot)` |
| stability | `classify.go` | `Classify(histories []TestHistory)` |
| suppression | `detect.go` | `Detect(snap *models.TestSuiteSnapshot)` |
| failure | `classify.go` | `Classify(inputs []FailureInput)` |
| clustering | `detect.go` | `Detect(snap *models.TestSuiteSnapshot)` |
| assertion | `assess.go` | `Assess(snap *models.TestSuiteSnapshot)` |
| envdepth | `assess.go` | `Assess(snap *models.TestSuiteSnapshot)` |
| changescope | `analyze.go` | `AnalyzePR(scope *impact.ChangeScope, snap *models.TestSuiteSnapshot)` |

## How to Add a New Assessment Dimension

Follow these steps to add a new analytical dimension (e.g., "test coupling", "setup complexity", "data dependency depth").

### Step 1: Create the package

```
internal/newdim/
  model.go
  assess.go        # or classify.go, detect.go -- pick the verb that fits
  newdim_test.go
```

### Step 2: Define the model (`model.go`)

Every model file follows this pattern:

```go
package newdim

// DimClass classifies the <dimension> of a test.
type DimClass string

const (
    ClassHigh    DimClass = "high"
    ClassLow     DimClass = "low"
    ClassUnknown DimClass = "unknown"   // Always include an unknown/unclear state
)

// Assessment is the <dimension> assessment for one test file.
type Assessment struct {
    FilePath    string
    Class       DimClass
    Confidence  float64    // 0.0-1.0, always present
    Explanation string     // Human-readable, always present
    // Domain-specific fields...
}

// AssessmentResult holds all <dimension> assessments.
type AssessmentResult struct {
    Assessments  []Assessment
    ByClass      map[DimClass]int
    OverallClass DimClass
}
```

Rules:
- Use typed string constants for enums, not raw strings or iota.
- Always include an "unknown", "unclear", or "data_insufficient" state.
- Every result struct must carry `Confidence float64` (0.0-1.0) and `Explanation string`.
- Use `map[EnumType]int` for aggregate counts, not parallel fields (exception: suppression uses named count fields for its 4 kinds, which is acceptable for small fixed sets).

### Step 3: Implement the logic

```go
package newdim

import (
    "sort"
    "github.com/pmclSF/hamlet/internal/models"
)

// Assess evaluates <dimension> across all test files in the snapshot.
func Assess(snap *models.TestSuiteSnapshot) *AssessmentResult {
    result := &AssessmentResult{
        ByClass: make(map[DimClass]int),
    }

    if snap == nil || len(snap.TestFiles) == 0 {
        result.OverallClass = ClassUnknown
        return result
    }

    for _, tf := range snap.TestFiles {
        a := assessFile(tf)
        result.Assessments = append(result.Assessments, a)
        result.ByClass[a.Class]++
    }

    // Sort for determinism.
    sort.Slice(result.Assessments, func(i, j int) bool {
        return result.Assessments[i].FilePath < result.Assessments[j].FilePath
    })

    result.OverallClass = computeOverall(result.ByClass)
    return result
}
```

Rules:
- Accept `*models.TestSuiteSnapshot` or typed input structs. Never accept raw `map[string]any` or untyped data.
- Handle nil/empty inputs gracefully -- return a valid result with unknown/insufficient states, not an error.
- Sort all output slices for deterministic results. Use `sort.Slice` with stable comparison keys.
- Keep subsystems independent. Do not import from other assessment packages.
- Only import from `internal/models`, `internal/signals`, `internal/impact`, or the standard library.

### Step 4: Write tests (`newdim_test.go`)

```go
package newdim

import (
    "testing"
    "github.com/pmclSF/hamlet/internal/models"
)

func TestAssess_NilSnapshot(t *testing.T) {
    result := Assess(nil)
    if result.OverallClass != ClassUnknown {
        t.Errorf("OverallClass = %s, want unknown", result.OverallClass)
    }
}

func TestAssess_EmptySnapshot(t *testing.T) {
    snap := &models.TestSuiteSnapshot{}
    result := Assess(snap)
    if result.OverallClass != ClassUnknown {
        t.Errorf("OverallClass = %s, want unknown", result.OverallClass)
    }
}

func TestAssess_HappyPath(t *testing.T) {
    snap := &models.TestSuiteSnapshot{
        TestFiles: []models.TestFile{
            {Path: "a_test.go", TestCount: 5, AssertionCount: 15},
            {Path: "b_test.go", TestCount: 3, AssertionCount: 1},
        },
    }
    result := Assess(snap)

    if len(result.Assessments) != 2 {
        t.Fatalf("len(Assessments) = %d, want 2", len(result.Assessments))
    }
    // Verify deterministic ordering.
    if result.Assessments[0].FilePath != "a_test.go" {
        t.Errorf("first assessment path = %s, want a_test.go", result.Assessments[0].FilePath)
    }
    // Verify confidence is set.
    for _, a := range result.Assessments {
        if a.Confidence <= 0 || a.Confidence > 1.0 {
            t.Errorf("confidence for %s = %f, want (0, 1]", a.FilePath, a.Confidence)
        }
        if a.Explanation == "" {
            t.Errorf("explanation for %s is empty", a.FilePath)
        }
    }
}

func TestAssess_DeterministicOutput(t *testing.T) {
    snap := buildFixture() // build a snapshot with multiple test files
    r1 := Assess(snap)
    r2 := Assess(snap)

    if len(r1.Assessments) != len(r2.Assessments) {
        t.Fatal("non-deterministic output length")
    }
    for i := range r1.Assessments {
        if r1.Assessments[i].FilePath != r2.Assessments[i].FilePath {
            t.Errorf("non-deterministic order at index %d", i)
        }
        if r1.Assessments[i].Class != r2.Assessments[i].Class {
            t.Errorf("non-deterministic classification at index %d", i)
        }
    }
}
```

Test patterns to follow:
- Test nil input, empty input, single-item input, multi-item input.
- Verify deterministic output ordering.
- Verify that every result has a non-zero confidence and a non-empty explanation.
- Build test fixtures inline or in helper functions within the test file (not in separate fixture files).
- Use `t.Errorf` for non-fatal assertions, `t.Fatalf` only when continuation is meaningless.

### Step 5: Integrate

Integration depends on where the new dimension should surface:

**In measurements** (e.g., feeding into posture dimensions):
- Add a new measurement to `internal/measurement/` using `internal/measurement/registry.go`.
- See `docs/contributing/adding-a-measurement.md` for the full measurement integration guide.

**In portfolio findings** (e.g., surfacing actionable insights):
- Add detection logic to the portfolio builder that consumes your assessment result.

**In the `hamlet show` command** (entity drill-down):
- See "How to Add a New CLI Drill-Down Entity" below.

**In `hamlet pr` output** (change-scoped analysis):
- Add finding generation in `changescope.buildChangeScopedFindings()`.

## How to Add a New CLI Drill-Down Entity

The `hamlet show` command supports entity types: `test`, `unit`, `owner`, `finding`. To add a new entity:

### Step 1: Add the case in `cmd/hamlet/main.go`

In the `runShow` function, add a new case to the switch:

```go
case "newentity":
    return showNewEntity(id, snap, jsonOutput)
```

### Step 2: Implement the show function

```go
func showNewEntity(id string, snap *models.TestSuiteSnapshot, jsonOutput bool) error {
    // Search for the entity in the snapshot.
    // ...

    if jsonOutput {
        enc := json.NewEncoder(os.Stdout)
        enc.SetIndent("", "  ")
        return enc.Encode(entity)
    }

    // Human-readable output.
    fmt.Printf("Entity: %s\n", entity.Name)
    fmt.Printf("... details ...\n")

    // Always suggest a next step.
    fmt.Println("\nNext: hamlet show test <path>   drill into a related test")
    return nil
}
```

Rules:
- Support both `--json` and human-readable output.
- Search by multiple identifiers (path, name, ID) to be forgiving of user input.
- Return a clear error message if the entity is not found: `fmt.Errorf("newentity not found: %s", id)`.
- Include a "Next:" hint suggesting a related drill-down command.

### Step 3: Update the usage string

In `printUsage()`, add the new entity to the `show` line:

```go
fmt.Fprintln(os.Stderr, "  show <entity> <id>       drill into test, unit, owner, finding, or newentity")
```

Also update the error message in `runShow`:

```go
return fmt.Errorf("unknown entity type: %q (valid: test, unit, owner, finding, newentity)", entity)
```

## How to Add a New Output Format

The `hamlet pr` command supports output formats via `--format`. To add a new format:

### Step 1: Add a renderer in `internal/changescope/render.go`

Follow the existing pattern:

```go
// RenderNewFormat writes output in the new format.
func RenderNewFormat(w io.Writer, pr *PRAnalysis) {
    // Write to w, not to os.Stdout directly.
    // This enables testing with bytes.Buffer.
    fmt.Fprintf(w, "... formatted output ...\n")
}
```

Rules:
- Accept `io.Writer`, not `*os.File`. This makes the renderer testable.
- Limit output lists (findings, recommended tests) to a reasonable cap (10-20 items) with a "... and N more" overflow.
- Include limitations in a collapsible section or footer.

### Step 2: Wire it in `cmd/hamlet/main.go`

In `runPR`, add the format case:

```go
case "newformat":
    changescope.RenderNewFormat(os.Stdout, pr)
```

### Step 3: Update usage

Add the format to the PR-specific flags section in `printUsage()`:

```go
fmt.Fprintln(os.Stderr, "  --format FORMAT          output: markdown, comment, annotation, newformat")
```

## How to Keep Assessments Cautious

Hamlet's assessments are designed to be honest about their limitations. Follow these principles:

### Always include an unknown/unclear state

Every classification enum must have a fallback for when evidence is insufficient:

```go
const (
    ClassStrong  MyClass = "strong"
    ClassWeak    MyClass = "weak"
    ClassUnclear MyClass = "unclear"   // Required: the "I don't know" state
)
```

### Use confidence metadata

Every assessment carries a `Confidence float64` field (0.0-1.0). Set it honestly:

| Confidence Range | Meaning | When to Use |
|-----------------|---------|-------------|
| 0.85-1.0 | High confidence | Exact match, unambiguous evidence |
| 0.60-0.84 | Moderate confidence | Pattern-based inference, heuristic match |
| 0.30-0.59 | Low confidence | Weak signals, partial evidence |
| 0.0-0.29 | Very low | Fallback classification, no clear evidence |

### Write honest explanations

Explanations should describe what was observed, not what was concluded:

```go
// Good: describes observation
"high retry rate (0.45) suggests retry-as-policy pattern"
"3 slow tests share code unit \"db/connection.go\""
"fewer than 3 observations available"

// Bad: over-claims
"this test is definitely flaky"
"root cause identified as db/connection.go"
"test suite is unhealthy"
```

### State limitations explicitly

When an assessment cannot be made, say why:

```go
if len(presentObs) < MinHistoryDepth {
    c.Class = ClassDataInsufficient
    c.Confidence = 0.3
    c.Explanation = "fewer than 3 observations available"
    return c
}
```

### Prefer false negatives over false positives

It is better to report "unclear" than to misclassify. Use conservative thresholds:

- Minimum cluster size: 3 (not 2)
- Minimum similarity score: 0.4 (not 0.2)
- Minimum history depth: 3 snapshots (not 1)

## Style Guide for Signal Explanations and Suggested Actions

### Explanations

Explanations appear in CLI output, JSON responses, and PR comments. They should be:

- **Factual**: describe what was observed, not what it means
- **Specific**: include numbers, file paths, framework names
- **Concise**: one sentence, under 120 characters when possible
- **Lowercase start**: explanations are sentence fragments, not standalone sentences

Examples:

```
"error message contains assertion keywords: expect(result).toBe(42)"
"5 test files link to code unit \"src/db/pool.go\""
"test was stable but has recently started failing"
"file path contains suppression indicator: quarantine"
"high assertion density (4.2/test) with low mock ratio"
"Framework \"playwright\" implies browser-backed execution environment."
```

### Suggested actions

Suggested actions appear in changescope findings and portfolio recommendations. They should be:

- **Actionable**: start with a verb (add, remove, split, investigate, consider)
- **Specific**: name the file, test, or unit to act on
- **Non-prescriptive**: use "consider" for judgment calls, imperative for clear fixes

Examples:

```
"Add unit tests for AuthService before merging."
"Consider splitting this test file; 45 tests exceed recommended density."
"Investigate shared dependency db/pool.go as a potential root cause for 8 flaky tests."
"Remove quarantine marker if the underlying issue has been resolved."
```

## Test Fixture Patterns and the testdata Package

### Inline fixtures

For assessment subsystem tests, build fixtures inline:

```go
func TestClassify_ChronicallyFlaky(t *testing.T) {
    h := TestHistory{
        TestID:   "test-1",
        TestName: "should handle auth",
        Observations: []Observation{
            {SnapshotIndex: 0, Present: true, FlakySignal: true},
            {SnapshotIndex: 1, Present: true, FlakySignal: true},
            {SnapshotIndex: 2, Present: true, FlakySignal: true},
            {SnapshotIndex: 3, Present: true, FlakySignal: false},
        },
    }
    result := classifyOne(h)
    if result.Class != ClassChronicallyFlaky {
        t.Errorf("class = %s, want chronically_flaky", result.Class)
    }
}
```

### The `internal/testdata` package

The `internal/testdata/` package contains integration-level tests that exercise the full pipeline. It includes:

- `determinism_test.go` -- verifies that repeated analysis produces identical output
- `schema_test.go` -- validates snapshot JSON schema compliance
- `adversarial_test.go` -- tests against adversarial inputs (empty files, binary files, huge directories)
- `bench_test.go` -- benchmarks for analysis performance
- `cli_test.go` -- CLI command integration tests
- `sample-repo/` -- a synthetic repository used as test input

When adding a new subsystem, consider adding a case to `determinism_test.go` that verifies your subsystem produces identical output on repeated runs.

### Test fixture construction helpers

If you need a snapshot with specific properties, build it in the test:

```go
func makeSnapshot(testFiles ...models.TestFile) *models.TestSuiteSnapshot {
    return &models.TestSuiteSnapshot{
        TestFiles: testFiles,
    }
}

func makeTestFile(path string, tests, assertions, mocks int) models.TestFile {
    return models.TestFile{
        Path:           path,
        TestCount:      tests,
        AssertionCount: assertions,
        MockCount:      mocks,
    }
}
```

Do not create separate fixture files or shared test utility packages. Keep test helpers in the test file that uses them.
