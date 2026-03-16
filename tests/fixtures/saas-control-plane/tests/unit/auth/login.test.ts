import { describe, it, expect } from 'vitest';
import { authenticate, validateToken, refreshToken } from '../../../src/auth/login';

describe('authenticate', () => {
  it('should return token for valid credentials', () => {
    const r = authenticate('admin@acme.com', 'secret');
    expect(r.token).toContain('tok_');
    expect(r.expiresIn).toBe(3600);
  });
  it('should throw for empty email', () => {
    expect(() => authenticate('', 'pass')).toThrow('Missing credentials');
  });
});

describe('validateToken', () => {
  it('should accept valid token', () => { expect(validateToken('tok_x')).toBe(true); });
  it('should reject bad token', () => { expect(validateToken('bad')).toBe(false); });
});

describe('refreshToken', () => {
  it('should refresh valid token', () => {
    expect(refreshToken('tok_x').expiresIn).toBe(7200);
  });
});
