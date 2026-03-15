import { describe, it, expect } from 'vitest';
import { login } from '../../src/auth/login.js';

describe('legacy login tests', () => {
  it.skip('should handle SSO login', async () => {
    // TODO: SSO not implemented yet
    const user = await login('sso@test.com', 'sso-token');
    expect(user).toBeDefined();
  });

  it.skip('should handle LDAP login', async () => {
    // TODO: LDAP connector not available
    const user = await login('ldap@test.com', 'ldap-pass');
    expect(user).toBeDefined();
  });

  it.skip('should handle OAuth token refresh', async () => {
    // Skipped: flaky in CI
    expect(true).toBe(true);
  });
});
