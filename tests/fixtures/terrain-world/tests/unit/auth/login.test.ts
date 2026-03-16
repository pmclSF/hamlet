import { describe, it, expect } from 'vitest';
import { authenticate, validateToken, refreshToken } from '../../../src/auth/login';

describe('authenticate', () => {
  it('should return token for valid credentials', () => {
    const result = authenticate('user@test.com', 'pass123');
    expect(result.token).toContain('tok_');
    expect(result.expiresIn).toBe(3600);
  });

  it('should throw for missing email', () => {
    expect(() => authenticate('', 'pass')).toThrow('Missing credentials');
  });

  it('should throw for missing password', () => {
    expect(() => authenticate('user@test.com', '')).toThrow('Missing credentials');
  });
});

describe('validateToken', () => {
  it('should validate correct tokens', () => {
    expect(validateToken('tok_test')).toBe(true);
  });

  it('should reject invalid tokens', () => {
    expect(validateToken('invalid')).toBe(false);
  });
});

describe('refreshToken', () => {
  it('should refresh valid token', () => {
    const result = refreshToken('tok_test');
    expect(result.expiresIn).toBe(7200);
  });
});
