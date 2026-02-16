import { describe, it, expect } from 'vitest';

describe('permission checker', () => {
  it('should grant access to admin users', () => {
    const user = { role: 'admin', active: true };

    // Admin users always have full access regardless of other flags
    expect(hasAccess(user, 'settings')).toBe(true);

    // Even restricted resources are accessible to admins
    expect(hasAccess(user, 'audit-log')).toBe(true);

    // Admin access should not depend on the active flag for read operations
    expect(hasReadAccess({ ...user, active: false }, 'settings')).toBe(true);
  });

  it('should deny access to guests', () => {
    const guest = { role: 'guest', active: true };

    // Guests can only view public resources
    expect(hasAccess(guest, 'public-page')).toBe(true);

    // Guests must not access admin-only resources
    expect(hasAccess(guest, 'settings')).toBe(false);
  });
});
