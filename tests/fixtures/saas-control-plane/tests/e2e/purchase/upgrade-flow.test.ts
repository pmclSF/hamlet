import { describe, it, expect } from 'vitest';
import { authenticate } from '../../../src/auth/login';
import { createInvoice } from '../../../src/billing/invoice';
import { createSubscription, changeplan } from '../../../src/billing/subscription';
import { connectDB, seedOrg, seedUser, cleanupDB } from '../../../src/shared-db';

describe('upgrade e2e', () => {
  it('should complete upgrade flow', () => {
    connectDB(); seedOrg(); seedUser();
    authenticate('buyer@acme.com', 'pass');
    const sub = createSubscription('org_test', 'starter');
    changeplan(sub.subscriptionId, 'enterprise');
    createInvoice('org_test', 19900);
    cleanupDB();
  });
});
