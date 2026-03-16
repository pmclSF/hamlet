# Generated Ground Truth Fixtures

Three fixture repositories created to pressure-test Terrain's reasoning engine with precise ground truth specifications.

## Fixture Summary

| Fixture | Domains | Source Files | Test Files | Scenarios | Problems | Score | Status |
|---------|---------|-------------|-----------|-----------|----------|-------|--------|
| saas-control-plane | 7 | 15 | 16 | 3 | 8 | 100% | 6/6 PASS |
| python-ml-observatory | 6 | 15 | 15 | 5 | 7 | 100% | 6/6 PASS |
| legacy-omnichannel | 7 | 16 | 12 | 3 | 8 | 94% | 6/6 PASS |

## saas-control-plane

**Theme:** B2B SaaS platform with auth, billing, entitlements, audit, search, notifications, and AI assistant.

**Validation layers:** Unit, integration, e2e, contract, AI eval, manual coverage.

**Key problems tested:**
- Shared database fanout (10+ dependents)
- Duplicate admin onboarding e2e flows
- Duplicate purchase e2e flows
- Weak billing coverage (subscription untested)
- Manual-only notifications zone
- Overlapping AI safety scenarios
- Weak assertions in search tests

**Truth categories:** impact (11 checks), coverage (2), redundancy (2), fanout (1), ai (7), environment.

**Final evaluator result:** 100% F1, 6/6 passed.

## python-ml-observatory

**Theme:** Python ML platform with data loaders, classifier, embeddings, retrieval, prompt builder, batch scoring, and safety filters.

**Validation layers:** Unit (pytest), eval suites, scenario declarations.

**Key problems tested:**
- Duplicate classifier accuracy eval files
- Duplicate safety eval files
- Untested data augmentation module
- Untested batch scoring module
- Fanout via shared conftest_helpers module
- Overlapping AI accuracy scenarios (same surfaces)
- Overlapping AI safety scenarios (same surfaces)

**Truth categories:** impact (11 checks), coverage (2), redundancy (2), fanout (1), ai (7), environment.

**Final evaluator result:** 100% F1, 6/6 passed.

## legacy-omnichannel

**Theme:** Mixed JS/TS commerce repo with Jest, Mocha, Cypress-era legacy overlap, plus AI merchandising.

**Validation layers:** Unit (mixed CJS/ESM), integration, e2e, AI eval, manual coverage.

**Key problems tested:**
- Duplicate checkout unit tests
- Duplicate e2e checkout flows
- Uncovered refunds module (manual-only)
- Uncovered mobile module (manual-only)
- High-fanout shared DB helper
- Overlapping recommendation safety scenarios
- Mixed module systems (CJS cart + ESM checkout)
- Legacy framework overlap (package.json declares 4 frameworks)

**Truth categories:** impact (14 checks), coverage (2), redundancy (1), fanout (1), ai (6), environment.

**Final evaluator result:** 94% F1, 6/6 passed. Coverage precision is 50% due to 2 unexpected uncovered files (payment.ts and db-helper.ts) that are legitimate findings not in the truth spec.

## Omitted Categories

All three fixtures omit the **stability** truth category because skip/flaky detection requires runtime artifacts (`--runtime` flag with JUnit XML or Jest JSON data). Without runtime data, Terrain cannot produce stability signals, and the evaluator would report false failures.

## Verification Commands

```bash
go run ./cmd/terrain-truthcheck --root tests/fixtures/saas-control-plane --truth tests/fixtures/saas-control-plane/tests/truth/terrain_truth.yaml
go run ./cmd/terrain-truthcheck --root tests/fixtures/python-ml-observatory --truth tests/fixtures/python-ml-observatory/tests/truth/terrain_truth.yaml
go run ./cmd/terrain-truthcheck --root tests/fixtures/legacy-omnichannel --truth tests/fixtures/legacy-omnichannel/tests/truth/terrain_truth.yaml
```
