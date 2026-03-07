# Detector Audit: Evidence Classification

This document classifies every Hamlet signal detector by evidence strength,
source, and detection method. It serves as the canonical reference for
understanding signal credibility.

## Evidence Model

Each signal carries three evidence fields:

| Field | Values | Purpose |
|-------|--------|---------|
| `evidenceStrength` | `strong`, `moderate`, `weak` | How robust the evidence is |
| `evidenceSource` | `ast`, `structural-pattern`, `path-name`, `runtime`, `coverage`, `policy` | How the signal was derived |
| `confidence` | 0.0 - 1.0 | Numeric certainty score |

Reports use these fields to calibrate language: strong-evidence findings are
stated directly, while weak-evidence findings include appropriate caveats.

## Detector Classification

### Quality Detectors

| Detector | Signal Type | Strength | Source | Confidence | Method |
|----------|------------|----------|--------|------------|--------|
| WeakAssertion (zero assertions) | `weakAssertion` | moderate | structural-pattern | 0.8 | Regex assertion counting in test files |
| WeakAssertion (low ratio) | `weakAssertion` | weak | structural-pattern | 0.6 | Assertion-to-test ratio heuristic |
| MockHeavy | `mockHeavyTest` | moderate | structural-pattern | 0.7-0.8 | Regex mock/assertion counting |
| UntestedExport | `untestedExport` | weak | path-name | 0.5 | Filename stem matching between test and source |
| CoverageThreshold | `coverageThresholdBreak` | strong | coverage | 0.9 | Istanbul/nyc coverage summary file |

### Migration Detectors

| Detector | Signal Type | Strength | Source | Confidence | Method |
|----------|------------|----------|--------|------------|--------|
| DeprecatedPattern | `deprecatedTestPattern` | moderate | structural-pattern | 0.7 | Regex: done-callbacks, setTimeout, Enzyme, Sinon |
| DynamicTestGeneration | `dynamicTestGeneration` | moderate | structural-pattern | 0.6 | Regex: forEach/map/test.each in test blocks |
| CustomMatcher | `customMatcherRisk` | weak | structural-pattern | 0.5 | Regex: expect.extend, chai.use, custom assert |
| UnsupportedSetup | `unsupportedSetup` | moderate | structural-pattern | 0.6 | Regex: global setup, root hooks, Cypress commands |
| FrameworkMigration | `frameworkMigration` | strong | structural-pattern | 0.8 | Multiple unit-test frameworks in same repo |

### Health Detectors (runtime-backed)

| Detector | Signal Type | Strength | Source | Confidence | Method |
|----------|------------|----------|--------|------------|--------|
| SlowTest | `slowTest` | strong | runtime | 0.9 | JUnit XML / Jest JSON duration data |
| FlakyTest | `flakyTest` | moderate | runtime | 0.7-0.8 | Retry metadata or mixed pass/fail outcomes |
| SkippedTest | `skippedTest` | strong | runtime | 0.9 | Status = skipped in runtime artifacts |

### Governance Detectors

| Detector | Signal Types | Strength | Source | Confidence | Method |
|----------|-------------|----------|--------|------------|--------|
| Governance | `policyViolation`, `legacyFrameworkUsage`, `runtimeBudgetExceeded` | strong | policy | 1.0 | Policy rule evaluation against snapshot |

## Improvement Path

The evidence model supports a layered approach to improving signal quality:

1. **Current**: Most detectors use structural-pattern matching (regex on file content). This provides moderate-confidence signals for JavaScript/TypeScript test files.

2. **Near-term**: AST-backed extraction for JS/TS would upgrade structural-pattern detectors to strong evidence. The `evidenceSource` field would change from `structural-pattern` to `ast`.

3. **Multi-language**: As parser support expands to Java, Python, and Go test files, the same detector interfaces produce signals with appropriate evidence strength for each language.

4. **Runtime enrichment**: When runtime artifacts are supplied, health detectors produce strong-evidence signals. Without artifacts, health findings are absent rather than guessed.

The key design principle: **honest about what we know**. Weak-evidence signals are not suppressed — they are labeled so that consumers (CLI, extension, summaries) can calibrate their language accordingly.
