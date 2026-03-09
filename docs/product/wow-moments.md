# Wow Moments

These are the specific insights Hamlet should reliably produce that make users say "I didn't know that."

## Top 5 Wow Moments

### 1. Instability concentrated in a few persistent tests

**Insight:** "3 test files account for 80% of all flaky/unstable signals."

**Why it matters:** Teams often treat flakiness as a systemic problem. Showing it concentrates in a handful of files makes it actionable.

**Evidence chain:** flakyTest + unstableSuite signals → heatmap directory concentration → posture health dimension.

### 2. Public functions covered only by E2E

**Insight:** "12 exported functions in src/api/ have no unit test coverage — only E2E tests exercise them."

**Why it matters:** E2E tests are slow, brittle, and expensive. Discovering that core API surface relies entirely on E2E coverage reveals structural risk.

**Evidence chain:** untestedExport signals + e2e_concentration measurement + framework type classification.

### 3. Migration risk compounded by weak test quality

**Insight:** "src/auth/ has 5 migration blockers AND 3 quality issues. Address quality before migrating."

**Why it matters:** Migrating code with weak assertions and heavy mocking is dangerous — tests pass but verify nothing. This cross-referencing is Hamlet's unique value.

**Evidence chain:** migration blocker signals + quality signals → area assessment (risky) → coverage guidance.

### 4. One owner/area carries disproportionate risk

**Insight:** "Team Platform owns 60% of all high-risk signals but only 20% of test files."

**Why it matters:** Risk concentration is invisible without ownership-aware analysis. Surfacing it enables targeted investment.

**Evidence chain:** ownership resolution + signal-per-owner heatmap + risk surface scoring.

### 5. Framework fragmentation creating maintenance burden

**Insight:** "5 different test frameworks across 40 test files. Framework fragmentation ratio: 0.125."

**Why it matters:** Teams accumulate frameworks over time without noticing. Quantifying fragmentation makes the maintenance cost visible.

**Evidence chain:** framework detection → framework_fragmentation measurement → coverage_diversity posture.

### 6. Runtime dominated by overlapping broad tests

**Insight:** "3 E2E tests cover 90% the same modules, consuming 60% of CI runtime."

**Why it matters:** Teams add broad E2E tests without realizing they duplicate coverage. The portfolio view reveals redundancy candidates, high-leverage tests, and runtime concentration — turning CI cost into an optimization target.

**Evidence chain:** `hamlet portfolio` → test cost signals + protection breadth overlap → redundancy candidates + runtime concentration ratio.

## What makes these moments work

1. **Structural, not obvious.** These are insights that require cross-referencing multiple signals — not something a developer would notice by reading test files.

2. **Quantified, not vague.** "3 files cause 80% of flakiness" is actionable. "Your tests are flaky" is not.

3. **Action-oriented.** Each insight implies a clear next step — fix those 3 files, add unit tests for those exports, address quality before migrating.

4. **Honest about evidence.** When runtime data is missing, Hamlet says so. When evidence is partial, the posture reflects it.
