# Advanced Assessment and Workflow Architecture Map

This document maps all advanced assessment and workflow subsystems introduced in Stages 11-16. Each subsystem adds a new analytical dimension to Hamlet's test intelligence, operating on snapshot data and surfacing results through CLI commands, reports, and portfolio findings.

## Subsystem Inventory

| Subsystem | Package Path | Purpose | Key Types | Integration Points |
|-----------|-------------|---------|-----------|-------------------|
| Lifecycle | `internal/lifecycle` | Track test identity continuity across snapshots (exact match, rename, move, split, merge) | `ContinuityClass`, `ContinuityMapping`, `ContinuityResult`, `EvidenceBasis` | Consumes two `*models.TestSuiteSnapshot`; feeds stability classification and trend tracking |
| Stability | `internal/stability` | Classify tests into 7 historical stability classes based on longitudinal observations | `StabilityClass`, `TestHistory`, `Observation`, `Classification`, `ClassificationResult` | Consumes `[]TestHistory` built from ordered snapshots; feeds portfolio risk attribution |
| Suppression | `internal/suppression` | Detect quarantine, skip, expected-failure, and retry-wrapper suppression mechanisms | `SuppressionKind`, `SuppressionIntent`, `DetectionSource`, `Suppression`, `SuppressionResult` | Consumes `*models.TestSuiteSnapshot` signals, runtime stats, and file paths; feeds health measurements |
| Failure | `internal/failure` | Classify test failures into 8 taxonomy categories using pattern matching on error messages and stack traces | `FailureCategory`, `ClassificationConfidence`, `FailureClassification`, `TaxonomyResult`, `FailureInput` | Consumes `[]FailureInput` extracted from runtime artifacts; feeds clustering and portfolio findings |
| Clustering | `internal/clustering` | Detect common-cause clusters where shared helpers, fixtures, or imports drive broad instability or slowness | `ClusterType`, `Cluster`, `ClusterResult` | Consumes `*models.TestSuiteSnapshot` with signals and linked code units; feeds portfolio and remediation prioritization |
| Assertion | `internal/assertion` | Assess assertion strength per test file (density, categories, mock ratio) | `StrengthClass`, `AssertionCategory`, `Assessment`, `AssessmentResult` | Consumes `*models.TestSuiteSnapshot` test files; feeds coverage-depth measurements |
| Envdepth | `internal/envdepth` | Classify test environment depth (heavy mocking vs. real dependency vs. browser runtime) | `DepthClass`, `EnvironmentIndicator`, `Assessment`, `AssessmentResult` | Consumes `*models.TestSuiteSnapshot` test files and framework metadata; feeds coverage-diversity measurements |
| Changescope | `internal/changescope` | PR and change-scoped analysis workflows with multiple output formats | `PRAnalysis`, `ChangeScopedFinding`, `PostureDelta` | Consumes `*impact.ChangeScope` + `*models.TestSuiteSnapshot`; surfaces via `hamlet pr` command |

## Dependency and Data-Flow Diagram

```
                              Snapshot Sources
                    +---------------------------------+
                    |   hamlet analyze (pipeline)      |
                    |   --runtime JUnit/Jest JSON      |
                    |   --coverage LCOV/Istanbul        |
                    +---------+-----------------------+
                              |
                              v
                  +---------------------------+
                  |  models.TestSuiteSnapshot  |
                  |  (signals, test files,     |
                  |   test cases, code units,  |
                  |   runtime stats, ownership)|
                  +--+--+--+--+--+--+--+------+
                     |  |  |  |  |  |  |
         +-----------+  |  |  |  |  |  +----------+
         |              |  |  |  |  |              |
         v              |  |  |  |  v              v
  +------------+        |  |  |  | +----------+ +-------------+
  | suppression|        |  |  |  | | assertion| | envdepth    |
  | Detect()   |        |  |  |  | | Assess() | | Assess()    |
  +-----+------+        |  |  |  | +----+-----+ +------+------+
        |               |  |  |  |      |              |
        v               |  |  |  |      v              v
  SuppressionResult     |  |  |  | AssessmentResult  AssessmentResult
        |               |  |  |  |      |              |
        |               v  |  |  |      |              |
        |        +----------+ |  |      |              |
        |        | clustering| |  |      |              |
        |        | Detect()  | |  |      |              |
        |        +-----+----+ |  |      |              |
        |              |      |  |      |              |
        |              v      |  |      |              |
        |        ClusterResult |  |      |              |
        |              |      |  |      |              |
        |              |      v  v      |              |
        |              | +----------+   |              |
        |              | | failure  |   |              |
        |              | |Classify()|   |              |
        |              | +----+-----+   |              |
        |              |      |         |              |
        |              |      v         |              |
        |              | TaxonomyResult |              |
        |              |      |         |              |
        +-------+------+------+---------+--------------+
                |                    |
                v                    v
        +--------------+    +-----------------+
        | measurement  |    | portfolio       |
        | framework    |    | intelligence    |
        +--------------+    +-----------------+
                |                    |
                v                    v
        +--------------+    +------------------+
        | reporting    |    | hamlet summary   |
        | posture/     |    | hamlet portfolio |
        | executive    |    | hamlet show      |
        +--------------+    +------------------+

  Snapshot History (ordered list of snapshots)
        |
        v
  +------------+       +------------+
  | lifecycle  | ----> | stability  |
  | InferCont  |       | BuildHist  |
  | inuity()   |       | Classify() |
  +-----+------+       +-----+------+
        |                     |
        v                     v
  ContinuityResult    ClassificationResult
        |                     |
        +----------+----------+
                   |
                   v
           +--------------+
           | comparison   |
           | trend reports|
           +--------------+

  Git Diff (hamlet pr / hamlet impact)
        |
        v
  +-------------------+
  | impact.ChangeScope|
  +--------+----------+
           |
           v
  +-------------------+
  | changescope       |
  | AnalyzePR()       |
  +--------+----------+
           |
           v
  +-------------------+
  | PRAnalysis        |
  | (findings, gaps,  |
  |  recommendations) |
  +--------+----------+
           |
     +-----+-----+-----+
     |           |      |
     v           v      v
  markdown   annotation  comment
  report     CI output   concise
```

## How Each Subsystem Consumes Snapshot Data

### Lifecycle (`InferContinuity`)

Accepts two `*models.TestSuiteSnapshot` pointers (from/to). Indexes `TestCases` by `TestID` for O(1) exact matching in Phase 1. Phase 2 applies heuristic scoring for unmatched tests using:

- **Name similarity**: LCS-based string similarity on `TestName` (0.4 weight at >=0.8, 0.2 at >=0.5)
- **Suite hierarchy similarity**: LCS on joined `SuiteHierarchy` (0.2 weight at >=0.8)
- **Path similarity**: weighted combination of directory similarity (0.4) and filename similarity (0.6)
- **Canonical identity similarity**: LCS on `CanonicalIdentity` (0.2 weight at >=0.7)

Phase 3 detects split patterns (1:N name-prefix relationships in same directory) and merge patterns (N:1 name-prefix relationships). Minimum score threshold is 0.4.

### Stability (`BuildHistories` + `Classify`)

`BuildHistories` takes an ordered slice of snapshots (oldest first). For each unique `TestID` found across all snapshots, it constructs a `TestHistory` by walking each snapshot and recording an `Observation` with:

- Presence (does this test exist in this snapshot?)
- Pass/fail status (from `RuntimeStats.PassRate` thresholds: >=0.95 = passed, <0.5 = failed)
- Skip status (from `skippedTest` signals)
- Runtime data (`AvgRuntimeMs`, `RetryRate`)
- Signal data (flaky, slow signals on the test's file)

`Classify` evaluates each history through an ordered decision tree: skip rate >=0.5 -> quarantined; improving trend with problems -> improving; flaky rate >=0.4 -> chronically flaky; newly unstable (stable early, failing late) -> newly unstable; slow rate >=0.3 -> intermittently slow; low fail+flaky -> consistently stable. Falls through to `data_insufficient`.

### Suppression (`Detect`)

Operates on a single snapshot. Three detection strategies run in sequence:

1. **Signal-based**: scans `snap.Signals` for `skippedTest` signal type -> `KindSkipDisable` at 0.9 confidence
2. **Runtime-based**: scans `snap.TestFiles` for high retry rate (>=0.3 -> `KindRetryWrapper` at 0.7) or very low pass rate (<0.3 -> `KindExpectedFailure` at 0.5)
3. **Naming-based**: scans `snap.TestFiles` for quarantine/skip/disabled/pending patterns in file paths -> `KindQuarantined` or `KindSkipDisable` at 0.6

After detection, deduplicates by file+kind, then classifies intent: quarantined = chronic; high retry with >=0.5 rate = chronic; expected failure = chronic; all others = unknown.

### Failure (`Classify`)

Accepts `[]FailureInput` (decoupled from runtime artifact format). Each input provides `TestFilePath`, `TestName`, `ErrorMessage`, and `StackTrace`. The classifier combines error message and stack trace (lowercased), then checks 7 ordered pattern sets by priority:

1. Snapshot mismatch (priority 1, confidence 0.95)
2. Selector/UI fragility (priority 2, confidence 0.90)
3. Infrastructure/environment (priority 3, confidence 0.95)
4. Dependency/service failure (priority 4, confidence 0.90)
5. Timeout (priority 5, confidence 0.90)
6. Setup/fixture failure (priority 6, confidence 0.75)
7. Assertion failure (priority 7, confidence 0.80)

The lowest priority number wins when multiple patterns match. Unmatched failures receive `CategoryUnknown` with `ConfidenceWeak` at 0.2.

### Clustering (`Detect`)

Operates on a single snapshot. Five detection strategies each produce independent clusters:

1. **Shared imports** (`ClusterSharedImport`): groups test files by `LinkedCodeUnits`, clusters code units referenced by >= 3 test files
2. **Slow path** (`ClusterDominantSlowHelper`): groups slow-signal tests by shared code units, includes total runtime as impact metric
3. **Flaky fixture** (`ClusterDominantFlakyFixture`): groups flaky-signal tests by shared code units
4. **Setup path** (`ClusterGlobalSetupPath`): groups test files by (directory, signal type) pairs >= 3 files
5. **Repeated failure** (`ClusterRepeatedFailPattern`): groups snapshot-level signals by (directory, signal type) pairs >= 3 files

All strategies use `minClusterSize=3`. Confidence is computed per-strategy using clamped linear formulas. Results are sorted by affected count descending.

### Assertion (`Assess`)

Iterates `snap.TestFiles`. For each file with tests:

- Computes assertion density = `AssertionCount / TestCount`
- Infers categories: `SnapshotCount` -> `CategorySnapshot`, remainder -> `CategoryBehavioral`
- Classifies strength using ordered rules:
  - No assertions -> weak (0.8 confidence)
  - Mocks > assertions -> weak (0.7)
  - Snapshot ratio >=0.8 with density <2.0 -> weak (0.6)
  - E2E frameworks: adjusted thresholds (density >=2.0 = strong, >=1.0 = moderate)
  - Non-E2E: density >=3.0 with low mocks = strong, >=1.5 = moderate, <1.0 = weak

Overall strength: strong if >=50% strong and <20% weak; weak if >=50% weak; otherwise moderate.

### Envdepth (`Assess`)

Iterates `snap.TestFiles`. For each file, classifies based on ordered checks:

1. Browser framework (cypress, playwright, puppeteer, etc.) -> `DepthBrowserRuntime` (0.85)
2. `MockCount > 2*AssertionCount` or `MockCount >= 8` -> `DepthHeavyMocking` (0.80)
3. `MockCount > 0` and `MockCount <= AssertionCount` -> `DepthModerateMocking` (0.70)
4. Zero mocks with E2E/integration framework -> `DepthRealDependency` (0.65)
5. Default -> `DepthUnknown` (0.30)

Overall depth is the dominant class by count, with ties broken by environmental realism priority (browser > real > moderate > heavy > unknown).

### Changescope (`AnalyzePR`)

Accepts an `*impact.ChangeScope` (from git diff) and a snapshot. Workflow:

1. Delegates to `impact.Analyze(scope, snap)` for protection gap detection and test selection
2. Counts changed files by type (source vs. test)
3. Extracts recommended tests from impact result
4. Builds change-scoped findings from:
   - Protection gaps (with severity and suggested actions)
   - Existing signals on changed files
   - Untested exported units in the change area
5. Computes posture delta based on impact posture band
6. Constructs summary string

Four output renderers:
- `RenderPRSummaryMarkdown`: GitHub PR comment with posture badge, stats table, findings, recommended tests, affected owners, and limitations
- `RenderPRCommentConcise`: one-line badge + summary with high-severity count
- `RenderCIAnnotation`: GitHub Actions `::error`/`::warning`/`::notice` format per finding
- `RenderChangeScopedReport`: terminal-formatted report with sections

## Where Each Subsystem's Output Surfaces

| Subsystem | CLI Commands | Report Renderers | JSON Output |
|-----------|-------------|------------------|-------------|
| Lifecycle | `hamlet compare` | `reporting.RenderComparisonReport` | `hamlet compare --json` |
| Stability | `hamlet compare`, `hamlet summary` | Trend sections in executive summary | `hamlet compare --json` |
| Suppression | `hamlet analyze`, `hamlet posture` | Health dimension measurements | `hamlet analyze --json` (within signals/measurements) |
| Failure | `hamlet analyze`, `hamlet show finding` | Analyze report signal sections | `hamlet analyze --json` (within signals) |
| Clustering | `hamlet portfolio`, `hamlet summary` | Portfolio findings sections | `hamlet portfolio --json` |
| Assertion | `hamlet posture`, `hamlet show test` | Coverage-depth measurements | `hamlet posture --json` |
| Envdepth | `hamlet posture`, `hamlet show test` | Coverage-diversity measurements | `hamlet posture --json` |
| Changescope | `hamlet pr` | 4 renderers: markdown, comment, annotation, plain | `hamlet pr --json` |

## Known Limitations and Language/Runner Gaps

### Lifecycle

- Heuristic matching uses LCS-based string similarity with a 0.4 minimum threshold. Very short test names may produce false-positive matches.
- Split/merge detection requires name-prefix relationships and shared directory. Cross-directory splits are not detected.
- `EvidenceCoverageContinuity` basis exists in the model but is not yet wired to actual coverage data.
- Language-agnostic: operates on `TestCase` model fields, not language-specific AST. This is a strength (universal) and a limitation (no syntax-level rename tracking).

### Stability

- Requires `MinHistoryDepth` of 3 snapshots before classifying beyond `data_insufficient`. New projects will see limited value until history accumulates.
- Runtime-based observations depend on `--runtime` artifact ingestion. Without runtime data, pass/fail/skip status relies solely on signals, which may not be present.
- No language or runner-specific heuristics. Classification is purely observational.
- Trend analysis uses a simple early-half vs. late-half comparison. Gradual trends across many snapshots may not be detected accurately.

### Suppression

- Naming-convention detection is English-centric (patterns: "quarantine", "skip", "disabled", "pending", "xdescribe", "xit").
- Expected-failure detection from runtime data uses a coarse 0.3 pass-rate threshold. Tests with intentionally low pass rates for other reasons may be misclassified.
- No detection of language-specific suppression annotations (e.g., `@Disabled` in JUnit 5, `pytest.mark.skip`, `@Ignore`) beyond what the signal pipeline already surfaces.
- Intent classification (tactical vs. chronic) has no time dimension without snapshot history.

### Failure

- Pattern matching is case-insensitive substring matching. Ambiguous tokens (e.g., "expected" in non-assertion contexts) may produce false classifications.
- 7 pattern sets cover JavaScript/TypeScript, Java, and Python idioms. Go, Ruby, C#, and other ecosystems have weaker coverage.
- Stack trace analysis is pattern-based, not structural. Framework-wrapped or truncated stack traces may not match.
- The `CategoryUnknown` fallback means some failures will always remain unclassified.

### Clustering

- `minClusterSize` is hardcoded at 3. Small test suites may not surface any clusters.
- Shared-import clustering depends on `LinkedCodeUnits` being populated by the static analysis pipeline. If import resolution fails, no shared-import clusters will appear.
- No cross-snapshot clustering. Each analysis looks at one snapshot in isolation, so persistent common causes are not tracked over time.
- Confidence formulas are linear approximations, not statistically calibrated.

### Assertion

- Assertion counting comes from static analysis (`AssertionCount`, `MockCount`, `SnapshotCount` on `TestFile`). Dynamically generated assertions or custom assertion libraries may not be counted.
- Category inference is coarse: only snapshot vs. behavioral. Finer categories (existence, status, type, exception) exist in the model but are not populated from snapshot data alone.
- E2E threshold adjustment is binary (is/is not E2E). No graduated threshold for integration or hybrid frameworks.

### Envdepth

- Classification relies on framework name and mock/assertion counts. Custom test setups that use real dependencies without known framework names receive `DepthUnknown`.
- Environment indicators (fake clock, stubbed network, in-memory DB, real HTTP) exist in the model but are only populated via framework name matching, not code-level detection.
- Browser and integration framework lists are hardcoded and JavaScript/Java-centric. Python and Go test frameworks are underrepresented.

### Changescope

- Git diff resolution delegates to `impact.ChangeScopeFromGitDiff`. Merge commits and rebases may produce unexpected change sets.
- Posture delta has no historical baseline comparison -- only the current-snapshot posture band is evaluated.
- Markdown renderer limits findings and recommended tests to 10 items per section.
- No integration with GitHub/GitLab PR comment APIs (output-only; no automated posting).

## Future Extension Points

### Lifecycle

- Wire `EvidenceCoverageContinuity` by comparing coverage attribution across snapshots.
- Add AST-level similarity for rename detection (language-specific plugins).
- Support configurable similarity thresholds via `.hamlet/policy.yaml`.

### Stability

- Add time-weighted classification where recent observations carry exponentially more weight.
- Integrate lifecycle continuity mappings to track stability across renames and moves.
- Support owner-scoped stability reports (aggregate stability by team).

### Suppression

- Detect language-specific suppression annotations (`@Disabled`, `@pytest.mark.skip`, `xit/xdescribe`) via dedicated signal detectors.
- Add time-aware intent classification using snapshot history (how long has this been suppressed?).
- Surface suppression remediation recommendations in portfolio findings.

### Failure

- Expand pattern sets for Go (`testing.T` failures), Ruby (RSpec), and C# (NUnit/xUnit).
- Add structural stack trace analysis (frame counting, framework frame filtering).
- Feed failure taxonomy into clustering as an additional signal dimension.

### Clustering

- Cross-snapshot clustering to detect persistent common causes over time.
- Owner-scoped clustering (which team's tests cluster together?).
- Cluster remediation cost estimation based on affected test count and total runtime.

### Assertion

- Populate finer assertion categories from AST analysis (existence checks, type checks, exception assertions).
- Configurable strength thresholds via policy.
- Per-test-case assertion assessment (currently per-file only).

### Envdepth

- Detect environment indicators from code patterns (e.g., `jest.useFakeTimers()`, `nock`, `testcontainers`).
- Language-specific indicator detection for Java (Mockito, WireMock, Testcontainers) and Python (unittest.mock, responses, moto).
- Feed depth classification into coverage-diversity measurements as a weighted signal.

### Changescope

- Direct GitHub/GitLab PR comment posting via API integration.
- Baseline comparison: compare PR posture against the base branch posture, not just the absolute band.
- Configurable finding severity thresholds via policy.
- Monorepo-scoped analysis (restrict to changed package boundaries).
- Integration with lifecycle/stability/assertion/clustering assessments for enriched PR findings.
