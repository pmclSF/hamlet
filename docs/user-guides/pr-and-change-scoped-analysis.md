# PR and Change-Scoped Analysis

Terrain can analyze changes in the context of a PR or local diff, focusing on the affected area rather than the entire repo.

## Quick Start

```bash
# Analyze current changes against HEAD~1
terrain pr

# Analyze against a specific base branch
terrain pr --base origin/main

# Get markdown output for PR comments
terrain pr --format markdown

# Get concise one-liner for inline comments
terrain pr --format comment

# Get CI annotation output (GitHub Actions compatible)
terrain pr --format annotation

# JSON output for programmatic consumption
terrain pr --json
```

## What PR Analysis Shows

- **Change posture**: How well-protected the changed code is
- **Protection gaps**: Changed code lacking test coverage
- **Existing signals**: Quality or health issues on changed files
- **Untested exports**: New or modified public API without tests
- **Recommended tests**: Which tests to run for this change
- **Affected owners**: Which teams own the impacted code

## Output Formats

### Human-readable (default)

```
Terrain — Change-Scoped Analysis
========================================

Posture:   PARTIALLY_PROTECTED
Files:     3 changed (2 source, 1 test)
Units:     4 impacted
Gaps:      1

Findings
----------------------------------------
  [HIGH] src/feature.js — Exported NewFeature has no test coverage.

Recommended Tests
----------------------------------------
  src/__tests__/auth.test.js
  src/__tests__/feature.test.js

Affected Owners: team-platform, team-payments
```

### Markdown (`--format markdown`)

Produces a GitHub-compatible markdown summary suitable for PR comments.

### CI Annotations (`--format annotation`)

Produces `::error` and `::warning` lines compatible with GitHub Actions annotations.

## Relationship to `terrain impact`

- `terrain impact` is the underlying analysis engine with detailed drill-down views
- `terrain pr` wraps impact analysis with PR-oriented formatting and recommendations
- Use `terrain impact --show gaps` for deep-diving into protection gaps
- Use `terrain pr --format markdown` for automated PR comments

## Limitations

- Change detection uses `git diff`; uncommitted changes may not be captured
- Coverage lineage quality affects recommendation precision
- PR analysis shows change-area posture, not repo-wide posture
