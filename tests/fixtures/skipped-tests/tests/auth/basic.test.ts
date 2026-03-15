import { describe, it, expect } from 'vitest';
import { hashPassword, verifyPassword } from '../../src/auth/basic';

describe('Basic Auth', () => {
  it('should hash password deterministically', () => {
    const hash1 = hashPassword('mypassword');
    const hash2 = hashPassword('mypassword');
    expect(hash1).toBe(hash2);
  });

  it('should verify correct password', () => {
    const hash = hashPassword('mypassword');
    expect(verifyPassword('mypassword', hash)).toBe(true);
  });

  it('should reject incorrect password', () => {
    const hash = hashPassword('mypassword');
    expect(verifyPassword('wrongpassword', hash)).toBe(false);
  });
});
