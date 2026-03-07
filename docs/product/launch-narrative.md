# Launch Narrative

## The story

Every engineering team has tests. Most teams have no idea whether those tests are actually protecting them.

Line coverage says 80%. But:
- Half the exported functions have no unit tests
- The auth tests are flaky 30% of the time
- Three different test frameworks create maintenance burden
- Migration blockers hide in files with the weakest assertions

Traditional coverage tools count lines. Hamlet reads structure.

## The wedge: migration pain

The most common entry point is framework migration. A team decides to move from Jasmine to Jest, or Protractor to Playwright. They discover it is harder than expected.

Hamlet scans the repo and shows:
- How many migration blockers exist and where
- Which blockers are compounded by quality issues (weak assertions, heavy mocking)
- Which areas are safe to migrate first
- Where additional test coverage would most reduce migration risk

This is immediately valuable. No other tool provides this.

## The real product: test intelligence

Migration readiness is the wedge. The real product is structural test intelligence:

- **Posture dimensions** — health, coverage depth, coverage diversity, structural risk, operational risk
- **Evidence-based findings** — every measurement traces to concrete signals with explicit confidence
- **Risk concentration** — where risk clusters by directory and owner
- **Trend tracking** — snapshot comparison shows what improved and what regressed

## What's honest

- Hamlet is local-first and open source today
- Analysis is static (runtime enrichment is optional)
- Coverage data improves accuracy but is not required
- Measurements are honest about evidence gaps
- No hosted service yet — that is future work

## What's distinctive

- Signal-first architecture: findings are structured, typed, and traceable
- Measurement framework with explicit evidence strength
- Migration readiness with quality-aware area assessment
- Five posture dimensions, not one score
- Privacy-safe benchmark exports
