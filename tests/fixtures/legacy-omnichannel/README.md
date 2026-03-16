# legacy-omnichannel

Mixed JS/TS commerce platform with 7 domains, legacy framework overlap, and 8 intentional problems.

## Domains

| Domain | Source | Language | Purpose |
|--------|--------|----------|---------|
| cart | `src/cart/` | JS (CJS) | Shopping cart (legacy) |
| checkout | `src/checkout/` | TS (ESM) | Checkout, payments |
| fraud | `src/fraud/` | TS | Risk analysis |
| refunds | `src/refunds/` | TS | Refund processing (no tests) |
| mobile | `src/mobile/` | TS | Mobile orders (no tests) |
| recommendations | `src/recommendations/` | TS | AI merchandising |
| admin | `src/admin/` | TS | Dashboard, reporting |

## Intentional Problems

1. **Duplicate checkout unit tests** — `checkout.test.ts` and `checkout-v2.test.ts`
2. **Duplicate e2e checkout flows** — `full-checkout.test.ts` and `express-checkout.test.ts`
3. **Uncovered refunds module** — `src/refunds/refund.ts` has no tests
4. **Uncovered mobile module** — `src/mobile/app.ts` has no tests
5. **High-fanout shared helper** — `src/shared/db-helper.ts` imported by integration/e2e
6. **Overlapping AI scenarios** — recommendation-safety and safety-regression cover same surfaces
7. **Mixed module systems** — CJS cart + ESM checkout (migration pattern)
8. **Legacy framework overlap** — package.json declares jest, mocha, cypress, and vitest

## Omitted Truth Categories

- **stability** — no runtime artifacts
