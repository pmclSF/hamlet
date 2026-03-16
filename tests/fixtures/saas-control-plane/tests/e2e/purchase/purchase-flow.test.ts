import { describe, it, expect } from 'vitest';
import { authenticate } from '../../../src/auth/login';
import { createInvoice, finalizeInvoice } from '../../../src/billing/invoice';
import { createSubscription } from '../../../src/billing/subscription';
import { connectDB, seedOrg, seedUser, seedInvoice, cleanupDB } from '../../../src/shared-db';

describe('purchase e2e', () => {
  it('should complete purchase flow', () => {
    connectDB(); seedOrg(); seedUser();
    authenticate('buyer@acme.com', 'pass');
    const sub = createSubscription('org_test', 'pro');
    const inv = createInvoice('org_test', 9900);
    const finalized = finalizeInvoice(inv.invoiceId);
    expect(finalized.status).toBe('finalized');
    cleanupDB();
  });
});
