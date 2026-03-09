# Health Measurements

## Overview

Health measurements assess the reliability and maintenance state of the test suite itself. They feed the `health` posture dimension.

## Measurements

### health.flaky_share

**What:** Share of test files flagged as flaky or unstable.

**How:** Counts signals of type `flakyTest` and `unstableSuite`, divides by total test file count.

**Evidence:** Strong if runtime data is available (JUnit XML, Jest JSON). Weak if based on static analysis only, since flakiness detection without runtime data relies on structural heuristics.

**Thresholds:**
| Ratio | Band |
|-------|------|
| ≤ 5% | strong |
| ≤ 15% | moderate |
| ≤ 30% | weak |
| > 30% | critical |

**Why it matters:** Flaky tests erode team confidence in the suite. A high flaky share means CI results are noisy and developers may stop trusting failures.

**Limitations:** Without runtime data, the measurement is based on static analysis only (evidence: weak). Provide `--runtime` artifacts for stronger evidence.

### health.skip_density

**What:** Share of test files containing skipped tests.

**How:** Counts `skippedTest` signals, divides by total test file count.

**Evidence:** Always strong — skip detection is a static analysis finding.

**Thresholds:**
| Ratio | Band |
|-------|------|
| ≤ 5% | strong |
| ≤ 15% | moderate |
| ≤ 30% | weak |
| > 30% | critical |

**Why it matters:** Skipped tests are hidden maintenance debt. They represent assertions that were once important enough to write but are no longer being exercised.

### health.dead_test_share

**What:** Share of test files containing dead or unreachable tests.

**How:** Counts `deadTest` signals, divides by total test file count.

**Evidence:** Always strong.

**Thresholds:**
| Ratio | Band |
|-------|------|
| ≤ 2% | strong |
| ≤ 10% | moderate |
| ≤ 20% | weak |
| > 20% | critical |

**Why it matters:** Dead tests inflate coverage metrics and test counts without providing value. They make the suite harder to maintain and slower to run.

### health.slow_test_share

**What:** Share of test files flagged as slow.

**How:** Counts `slowTest` signals, divides by total test file count.

**Evidence:** Strong if runtime data is available. Weak if based on static heuristics only.

**Thresholds:**
| Ratio | Band |
|-------|------|
| ≤ 10% | strong |
| ≤ 25% | moderate |
| ≤ 50% | weak |
| > 50% | critical |

**Why it matters:** Slow tests lengthen CI feedback loops, discouraging developers from running the full suite. They are often integration tests that could be restructured.

## Test Fixtures

### Concentrated instability

```
Test files: 20
Flaky signals: 10 (all in src/auth/)
→ health.flaky_share = 0.50 (critical)
→ Health posture: critical
```

### Broad low-level instability

```
Test files: 100
Flaky signals: 8 (spread across 8 different directories)
→ health.flaky_share = 0.08 (moderate)
→ Health posture: moderate
```

### Many skipped tests

```
Test files: 50
Skipped signals: 20
→ health.skip_density = 0.40 (critical)
→ Health posture: critical
```

### Partial runtime data

```
Test files: 30, 10 with runtime stats
Flaky signals: 3
→ health.flaky_share = 0.10 (moderate), evidence: strong (some runtime data)
```

## File

`internal/measurement/health.go`
