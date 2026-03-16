# Terrain World — Comprehensive Benchmark Repository

A synthetic repository designed to pressure-test Terrain's reasoning engine across all analysis dimensions. Contains 7 business domains, 6 validation layers, and 12 intentional problems.

## Domains

| Domain | Source | Purpose |
|--------|--------|---------|
| auth | `src/auth/` | Authentication, sessions, MFA |
| payments | `src/payments/` | Charges, subscriptions |
| refunds | `src/refunds/` | Refund processing |
| fraud | `src/fraud/` | Transaction analysis, rules engine |
| notifications | `src/notifications/` | Email, push notifications |
| ai-assistant | `src/ai-assistant/` | Prompts, model inference, datasets |
| mobile | `src/mobile/` | Mobile purchase flow |

## Validation Layers

| Layer | Location | Count |
|-------|----------|-------|
| Unit tests | `tests/unit/` | ~25 tests across 7 files |
| Integration tests | `tests/integration/` | 3 tests across 3 files |
| E2E tests | `tests/e2e/` | 4 tests across 3 files |
| Contract tests | `tests/contract/` | 3 tests across 2 files |
| AI eval tests | `tests/eval/` | 8 tests across 4 files |
| Manual validations | `.terrain/terrain.yaml` | 4 manual coverage entries |
| AI scenarios | `.terrain/terrain.yaml` | 4 declared scenarios |

## Intentional Problems

### 1. High Fanout
`src/shared-db.ts` exports 12 helper functions imported by 6+ test files across integration and e2e layers. Any change to this file triggers a wide blast radius.

### 2. Weak Coverage
- `src/payments/subscription.ts` — 3 exported functions, zero tests
- `src/fraud/rules.ts` — 3 exported functions, zero direct test imports
- `src/payments/charge.ts` — only `createCharge` tested; `captureCharge` and `voidCharge` untested

### 3. Redundant Tests
- `tests/unit/mobile/purchase.test.ts` and `purchase-v2.test.ts` — near-identical test files
- `tests/e2e/purchase/full-purchase.test.ts` and `checkout-flow.test.ts` — overlapping e2e flows

### 4. Flaky Clusters (Skip Debt)
`tests/unit/refunds/refund.test.ts` has 4 skipped tests (`it.skip`) — skipped test burden of ~60% within that file.

### 5. Manual-Only Zones
`src/notifications/` (email + push) has zero automated test coverage. Coverage is declared via manual_coverage entries in `.terrain/terrain.yaml` only.

### 6. AI Scenario Duplication
Scenarios `prompt-safety` and `safety-regression` both cover the same 2 code surfaces (`buildSafetyPrompt`, `systemPrompt`). This is intentional overlap for scenario duplication detection testing.

### 7. Weak Assertions
`tests/unit/fraud/detector.test.ts` uses `toBeTruthy()` instead of specific matchers — weak assertion pattern that Terrain's quality detector should flag.

### 8. Environment Redundancy
Mobile tests (`tests/e2e/mobile/mobile-purchase.test.ts`) test both iOS and Android with nearly identical flows, creating platform-conditional duplication.

## Truth Spec

`tests/truth/terrain_truth.yaml` documents expected findings per analysis dimension:
- **impact** — expected test cascades from specific file changes
- **coverage** — expected uncovered and weakly-covered files
- **redundancy** — expected duplicate test clusters
- **fanout** — expected high-fanout nodes
- **stability** — expected skipped test patterns
- **ai** — expected scenario duplication and surface detection
- **environment** — expected platform coverage patterns

## Running

```bash
# Analyze
terrain analyze --root tests/fixtures/terrain-world

# Insights
terrain insights --root tests/fixtures/terrain-world

# AI list
terrain ai list --root tests/fixtures/terrain-world

# AI doctor
terrain ai doctor --root tests/fixtures/terrain-world
```
