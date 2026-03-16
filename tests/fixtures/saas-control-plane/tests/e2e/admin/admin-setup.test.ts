import { describe, it, expect } from 'vitest';
import { authenticate } from '../../../src/auth/login';
import { assignRole } from '../../../src/auth/rbac';
import { createSubscription } from '../../../src/billing/subscription';
import { logEvent } from '../../../src/audit/logger';
import { connectDB, seedOrg, seedUser, cleanupDB } from '../../../src/shared-db';

describe('admin setup e2e', () => {
  it('should set up admin account', () => {
    connectDB(); seedOrg(); seedUser();
    const auth = authenticate('admin@acme.com', 'pass');
    assignRole('usr_test', 'admin');
    createSubscription('org_test', 'pro');
    logEvent('admin', 'setup_complete', 'org_test');
    cleanupDB();
  });
});
