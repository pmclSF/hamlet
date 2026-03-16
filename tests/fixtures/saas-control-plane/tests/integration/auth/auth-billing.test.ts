import { describe, it, expect } from 'vitest';
import { authenticate } from '../../../src/auth/login';
import { checkPermission } from '../../../src/auth/rbac';
import { createInvoice } from '../../../src/billing/invoice';
import { connectDB, seedOrg, seedUser, cleanupDB } from '../../../src/shared-db';

describe('auth + billing integration', () => {
  it('should authenticate then create invoice', () => {
    connectDB(); seedOrg(); seedUser();
    const auth = authenticate('admin@acme.com', 'pass');
    const perm = checkPermission(auth.token, 'billing', 'write');
    expect(perm.allowed).toBe(true);
    const inv = createInvoice('org_1', 5000);
    expect(inv.status).toBe('draft');
    cleanupDB();
  });
});
