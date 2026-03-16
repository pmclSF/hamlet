import { describe, it, expect } from 'vitest';
import { authenticate } from '../../../src/auth/login';
import { assignRole } from '../../../src/auth/rbac';
import { createSubscription } from '../../../src/billing/subscription';
import { checkEntitlement } from '../../../src/entitlements/check';
import { logEvent } from '../../../src/audit/logger';
import { connectDB, seedOrg, seedUser, cleanupDB } from '../../../src/shared-db';

describe('admin onboarding e2e', () => {
  it('should complete full onboarding', () => {
    connectDB(); seedOrg(); seedUser();
    const auth = authenticate('admin@acme.com', 'pass');
    assignRole('usr_test', 'admin');
    createSubscription('org_test', 'enterprise');
    checkEntitlement('org_test', 'unlimited');
    logEvent('admin', 'onboarding_complete', 'org_test');
    cleanupDB();
  });
});
