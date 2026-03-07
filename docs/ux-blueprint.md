# UX Blueprint

Hamlet is a developer-first observability product for test systems.

The UX should feel:
- lightweight
- actionable
- trustworthy
- native to developer workflows

It should not feel like a heavy BI dashboard.

## Canonical user flow

Observe -> Understand -> Act -> Improve

### Observe
User runs `hamlet analyze` or opens the extension.

They should immediately understand:
- frameworks in use
- top health issues
- top quality issues
- top risk surfaces
- modernization readiness

### Understand
User drills into:
- Health
- Quality
- Migration
- Review

Every signal must explain:
- what was found
- why it matters
- what action is recommended

### Act
User:
- fixes weak tests
- reduces runtime hotspots
- improves coverage gaps
- handles blockers
- previews migration

### Improve
User or team:
- saves snapshots
- uses CI checks
- prevents regressions
- eventually compares risk trends

## Core UX surfaces

### CLI
Primary entry point.
Fast, compact, insight-rich.

### VS Code / Cursor
Thin IDE layer that renders snapshot data and signals.

Sidebar sections:
- Overview
- Health
- Quality
- Migration
- Review

### CI
Produces annotations and policy warnings.

## UX rules

### 1. Show examples, not just counts
Every summary should include representative files/examples.

### 2. Explain every score
No black-box quality or risk scores.

### 3. Use confidence behaviorally
High confidence can drive recommended actions.
Low confidence routes to review.

### 4. Triage by pattern first
Review should often group by blocker/signal type first, then by owner/package.

### 5. Avoid surveillance framing
Hamlet shows engineering risk surfaces, not individual rankings.
