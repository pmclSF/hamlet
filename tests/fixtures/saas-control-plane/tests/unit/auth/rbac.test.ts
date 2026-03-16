import { describe, it, expect } from 'vitest';
import { checkPermission, assignRole, listRoles } from '../../../src/auth/rbac';

describe('checkPermission', () => {
  it('should allow with valid token', () => {
    expect(checkPermission('tok_admin', 'billing', 'read').allowed).toBe(true);
  });
  it('should deny with invalid token', () => {
    expect(checkPermission('bad', 'billing', 'read').allowed).toBe(false);
  });
});

describe('assignRole', () => {
  it('should assign role', () => { expect(assignRole('u1', 'admin').assigned).toBe(true); });
});

describe('listRoles', () => {
  it('should return roles', () => { expect(listRoles('u1').length).toBeGreaterThan(0); });
});
