# Manual Coverage Overlays

Manual coverage lets Terrain acknowledge validation that happens outside CI — TestRail regression suites, QA checklists, exploratory testing sessions, and release sign-off procedures.

Manual coverage is an **overlay**: it supplements automated coverage reporting but is **never treated as executable CI validation**.

## Configuring Manual Coverage

Add a `manual_coverage` section to `terrain.yaml` in your repository root:

```yaml
manual_coverage:
  - name: billing regression suite
    area: billing-core
    source: testrail
    owner: qa-billing
    criticality: high
    frequency: per-release

  - name: onboarding flow
    area: onboarding
    source: jira
    owner: qa-platform
    criticality: high
    frequency: weekly
    last_executed: "2026-03-01"

  - name: admin portal smoke test
    area: admin
    source: checklist
    owner: qa-admin
    criticality: medium
    frequency: per-release
```

### Field Reference

| Field | Required | Default | Description |
|-------|----------|---------|-------------|
| `name` | yes | — | Human-readable label |
| `area` | yes | — | Code area covered (e.g., `billing-core`, `auth/login`, `checkout/*`) |
| `source` | no | `manual` | Origin system: `testrail`, `jira`, `qase`, `checklist`, `exploratory`, `manual` |
| `owner` | no | — | Team or individual responsible |
| `criticality` | no | `medium` | `high`, `medium`, or `low` |
| `frequency` | no | — | `per-release`, `weekly`, `monthly`, `ad-hoc` |
| `last_executed` | no | — | ISO 8601 date of last execution (used for staleness detection) |
| `surfaces` | no | — | Specific CodeSurface or BehaviorSurface IDs covered |

## Example: TestRail Integration

A team using TestRail for regression testing:

```yaml
manual_coverage:
  - name: Login & authentication regression
    area: auth
    source: testrail
    owner: qa-security
    criticality: high
    frequency: per-release
    last_executed: "2026-03-10"

  - name: Payment processing regression
    area: payments
    source: testrail
    owner: qa-payments
    criticality: high
    frequency: per-release
    last_executed: "2026-03-10"

  - name: Reporting dashboard validation
    area: reporting
    source: testrail
    owner: qa-analytics
    criticality: medium
    frequency: monthly
    last_executed: "2026-02-15"
```

## Example: QA Checklist Workflow

A team using checklists for release validation:

```yaml
manual_coverage:
  - name: Pre-release smoke test
    area: core
    source: checklist
    owner: release-team
    criticality: high
    frequency: per-release

  - name: Accessibility audit
    area: ui
    source: checklist
    owner: frontend-team
    criticality: medium
    frequency: monthly
    last_executed: "2026-02-01"
```

## Example: Exploratory Testing

A team with dedicated exploratory testing:

```yaml
manual_coverage:
  - name: New feature exploratory session
    area: checkout
    source: exploratory
    owner: senior-qa
    criticality: high
    frequency: weekly
    last_executed: "2026-03-12"

  - name: Edge case discovery
    area: billing-core
    source: exploratory
    owner: senior-qa
    criticality: medium
    frequency: ad-hoc
```

## How Manual Coverage Appears in Reports

### `terrain analyze`

```
Manual Coverage Overlay
------------------------------------------------------------
  Artifacts:  5 (not executable — supplements CI coverage)
  Sources:    testrail: 3, checklist: 1, exploratory: 1
  Criticality: high: 3, medium: 2
  Areas:      auth, billing-core, checkout, payments, reporting
  Stale:      1 artifact(s) have no recent execution date
```

### `terrain insights`

When manual coverage artifacts lack execution dates, insights flags this as coverage debt:

```
Coverage Debt (1)
------------------------------------------------------------
  [MEDIUM] 3 of 5 manual coverage artifacts have no recent execution date
           Stale manual coverage may provide false confidence.
           Verify these validation activities are still being performed.
```

### `terrain impact`

When a protection gap overlaps with a manually covered area:

```
Policy Notes:
  Manual coverage exists for auth: "Login & authentication regression"
    (testrail, high criticality). Not executable — verify manually.
```

## Key Distinction: Overlay vs. Executable

Manual coverage **does not**:
- Count toward automated test totals
- Satisfy executable coverage requirements
- Participate in CI test selection (`terrain impact`)
- Replace the need for automated tests

Manual coverage **does**:
- Appear in the repository profile (`manualCoveragePresence`)
- Annotate protection gaps in impact analysis
- Trigger staleness findings when not recently executed
- Inform risk assessment by acknowledging human QA processes
- Trigger the `LARGE_MANUAL_SUITE` edge case when significant
