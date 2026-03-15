# UI Regression Testing

Terrain's user interface is currently CLI-driven. Visual output is produced by
view-model transformations that convert analysis data into formatted reports.
UI regression tests validate these view-model functions to ensure reports
render correctly across all data states.

## View-Model Functions Under Test

Each render function transforms structured analysis data into a formatted
string suitable for terminal display:

| Function | Report Type | Key Content |
|----------|-------------|-------------|
| `RenderAnalyzeReport` | Full analysis | Signals, risk items, coverage, frameworks detected |
| `RenderSummaryReport` | Executive summary | Posture band, top risks, signal counts, recommendations |
| `RenderPostureReport` | Posture detail | Dimension scores, measurement breakdowns, band assignment |
| `RenderPortfolioReport` | Multi-repo overview | Cross-project posture comparison, aggregate metrics |
| `RenderImpactReport` | Impact analysis | Changed files, affected tests, protection status |
| `RenderImpactUnits` | Impacted code units | Code units affected by changes with coverage status |
| `RenderImpactGaps` | Protection gaps | Code changes lacking test coverage |
| `RenderImpactTests` | Affected tests | Tests that should run for the given changes |
| `RenderImpactGraph` | Dependency graph | Structural relationships between changed and tested code |
| `RenderProtectiveSet` | Protective tests | Tests that guard specific code paths |
| `RenderImpactOwners` | Ownership impact | Teams affected by changes with their test responsibilities |

## Test States

View-model tests cover four data states to ensure robust rendering:

### Empty State
No analysis data available. The renderer must produce a meaningful message
(e.g., "No test files found") rather than crashing or printing empty tables.
Every render function is tested with nil or zero-value input.

### Loading and Error States
Handled at the CLI layer rather than in view-model functions. The CLI
displays progress indicators during analysis and prints structured error
messages if the pipeline fails. View-model functions receive only
successfully computed data.

### Full Data
The primary test path. Fixtures with complete analysis results are rendered
and checked for:
- Correct section headers and labels
- Accurate numeric values (signal counts, percentages, scores)
- Proper table formatting and alignment
- All expected data rows present

### Filtered Data
When the user applies filters (by team, by directory, by framework), the
rendered output must reflect only the filtered subset. Tests verify that
filtering does not corrupt aggregations or produce misleading totals.

## Confidence Cues

Reports include visual indicators that help users interpret results:

### Posture Bands
Quality posture is bucketed into named bands (e.g., "Strong", "Moderate",
"Weak"). Tests verify that the correct band name appears for given scores
and that band thresholds are applied consistently.

### Impact Confidence Levels
Impact analysis reports a confidence level for each finding based on the
completeness of available data (dependency graph presence, coverage data
quality). Tests verify that confidence labels match the underlying data.

### Protection Status
Each code unit is marked as protected (has covering tests), partially
protected, or unprotected. Tests verify that status indicators render
correctly and that counts are consistent with the underlying data.

## Adding a New View-Model Test

1. Identify the render function to test in `internal/report/` or
   `internal/impact/`.
2. Create a fixture with the specific data state you want to validate.
3. Call the render function and assert on the output string:
   - Use `strings.Contains` for content presence checks.
   - Use exact string matching sparingly (only for critical labels).
   - Check absence of internal/debug fields.
4. Cover empty state, full data, and at least one filtered case.

## Relationship to E2E Tests

View-model tests validate rendering logic in isolation. E2E tests
(`e2e_test.go`) validate the full pipeline including rendering. If a
view-model test passes but the E2E test fails, the issue is in data
preparation rather than rendering.

See also:
- `docs/engineering/e2e-scenario-testing.md` for full pipeline tests
- `docs/user-guides/output-modes.md` for user-facing output documentation
- `docs/engineering/verification-system-map.md` for the full test layer diagram
