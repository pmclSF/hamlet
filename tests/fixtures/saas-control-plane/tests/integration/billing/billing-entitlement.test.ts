import { describe, it, expect } from 'vitest';
import { createSubscription } from '../../../src/billing/subscription';
import { checkEntitlement } from '../../../src/entitlements/check';
import { connectDB, seedOrg, cleanupDB } from '../../../src/shared-db';

describe('billing + entitlement integration', () => {
  it('should create subscription and check entitlement', () => {
    connectDB(); seedOrg();
    const sub = createSubscription('org_1', 'pro');
    expect(sub.status).toBe('active');
    const ent = checkEntitlement('org_1', 'api_calls');
    expect(ent.entitled).toBe(true);
    cleanupDB();
  });
});
