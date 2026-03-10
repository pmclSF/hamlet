# Impact Analysis Model

## Stage 121 -- Product Concept

### What Is Impact Analysis?

Impact analysis answers a deceptively simple question: **given a code change, which tests matter?**

Hamlet's impact analysis engine examines a set of changed files, resolves the code units they contain (functions, classes, modules), maps those units to the tests that exercise them, and surfaces the result as a structured impact report. The report tells you what changed, what protects it, and where protection is missing.

### Key Questions It Answers

- **Which tests matter for my change?** Rather than running the full suite, impact analysis identifies the protective test set -- the minimal collection of tests that exercise the changed code.
- **Where are protection gaps?** Changed code units that lack any mapped test are flagged as gaps. These are the areas most likely to harbor undetected regressions.
- **Is this change risky?** Impact analysis produces a change-risk posture that combines the volume of change, the proportion of gaps, and the confidence of test-to-code mappings into a single risk assessment.

### How It Fits With Existing Hamlet Capabilities

Impact analysis builds on three foundational layers already present in Hamlet's current engine:

| Layer | Role in Impact Analysis |
|-------|------------------------|
| **Signals** | Quality signals (coverage, duplication, flakiness) feed confidence scores for test-to-code mappings. |
| **Posture** | The migration/quality posture provides baseline risk context. Impact analysis adds change-specific risk on top. |
| **Portfolio** | Portfolio-level views aggregate impact across packages and owners, enabling org-wide triage. |

Impact analysis is not a replacement for posture or portfolio -- it is a focused lens applied to a specific diff. It reuses the same signal taxonomy and scoring model.

### Use Cases

**PR Review.** A developer opens a pull request. Hamlet analyzes the diff, identifies impacted tests, and posts a summary comment listing protective tests and gaps. Reviewers can focus attention on under-tested areas.

**CI Gating.** A CI pipeline runs `hamlet impact` against the PR diff. If the change-risk posture exceeds a threshold (e.g., high-risk with multiple gaps), the pipeline can block merge or require additional review.

**Selective Test Execution.** Instead of running the full test suite on every commit, `hamlet select-tests` outputs the protective test set for the current diff. CI runners execute only those tests, reducing feedback time without sacrificing coverage of changed code.

**Owner Notification.** Impact analysis maps changed units to CODEOWNERS. When a change crosses ownership boundaries, Hamlet identifies affected owners so they can be notified or added as reviewers.

### Limitations

- **Heuristic-based mapping.** Test-to-code mappings are inferred from static analysis (imports, naming conventions, co-location patterns). They are not derived from runtime execution traces. Some mappings will be approximate.
- **No runtime awareness without artifacts.** Without coverage reports or execution logs as input, Hamlet relies entirely on structural analysis. Providing coverage artifacts improves mapping confidence.
- **Inferred vs. exact confidence.** Each mapping carries a confidence level (exact or inferred). Inferred mappings are useful but may include false positives. The drill-down views let users filter by confidence.
- **Language scope.** Test-to-code mapping quality depends on the language and framework. Well-structured projects with conventional naming yield better results.
- **Transitive impact.** The engine traces direct imports and one level of transitive dependency. Deeply indirect impacts (e.g., a utility change affecting a service three layers up) may not be captured without explicit dependency graph input.

### Design Principles

1. **Actionable over exhaustive.** Surface the 20% of information that drives 80% of decisions.
2. **Confidence-aware.** Never present inferred mappings as certainties. Always show confidence level.
3. **Composable.** Impact analysis output is JSON-first, designed to feed CI tools, PR bots, and UI dashboards.
4. **Non-blocking by default.** Impact analysis informs; it does not block unless the user configures thresholds.
