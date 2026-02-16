import { describe, it, expect } from 'vitest';

const ROLES = ['admin', 'user', 'guest'] as const;
type Role = typeof ROLES[number];

describe('Roles', () => {
  it('should validate role', () => {
    const role: Role = 'admin';
    expect(ROLES).toContain(role);
  });

  it('should have three roles', () => {
    expect(ROLES.length).toBe(3);
  });

  it('should include guest', () => {
    const guest: Role = 'guest';
    expect(ROLES).toContain(guest);
  });
});
