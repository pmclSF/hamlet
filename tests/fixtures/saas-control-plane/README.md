# saas-control-plane

B2B SaaS platform fixture with 7 domains, 5 validation layers, and 8 intentional problems.

## Domains

| Domain | Source | Purpose |
|--------|--------|---------|
| auth | `src/auth/` | Login, RBAC, token management |
| billing | `src/billing/` | Invoices, subscriptions |
| entitlements | `src/entitlements/` | Feature gating, usage limits |
| audit | `src/audit/` | Event logging, export |
| search | `src/search/` | Document indexing |
| notifications | `src/notifications/` | Email, webhooks (manual-only) |
| ai-assistant | `src/ai-assistant/` | Prompts, model, datasets |

## Validation Layers

- Unit tests: auth, billing, entitlements, audit, search
- Integration tests: auth+billing, billing+entitlements
- E2E tests: admin onboarding, purchase, upgrade
- Contract tests: billing API
- AI eval: safety, accuracy
- Manual coverage: notifications, audit export

## Intentional Problems

1. **High fanout** — `src/shared-db.ts` imported by all integration and e2e tests
2. **Duplicate admin tests** — `admin-onboarding.test.ts` and `admin-setup.test.ts`
3. **Duplicate purchase tests** — `purchase-flow.test.ts` and `upgrade-flow.test.ts`
4. **Weak billing coverage** — subscription.ts has no tests; invoice.ts only tests createInvoice
5. **Manual-only notifications** — email.ts and webhook.ts have no automated tests
6. **Overlapping AI safety scenarios** — prompt-safety and safety-regression cover same surfaces
7. **Weak assertions** — search/index.test.ts uses toBeTruthy()
8. **Uncovered exports** — finalizeInvoice, voidInvoice, exportAuditLog untested

## Omitted Truth Categories

- **stability** — no runtime artifacts; skip detection requires `--runtime` data
