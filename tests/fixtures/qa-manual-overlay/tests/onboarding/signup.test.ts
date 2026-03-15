import { describe, it, expect } from 'vitest';
import { validateSignup, normalizeEmail } from '../../src/onboarding/signup';

describe('Signup', () => {
  it('should accept valid signup data', () => {
    const errors = validateSignup({
      email: 'user@example.com',
      password: 'securepass',
      name: 'Test User',
      acceptTerms: true,
    });
    expect(errors).toHaveLength(0);
  });

  it('should reject missing terms acceptance', () => {
    const errors = validateSignup({
      email: 'user@example.com',
      password: 'securepass',
      name: 'Test User',
      acceptTerms: false,
    });
    expect(errors).toContain('must accept terms');
  });

  it('should normalize email', () => {
    expect(normalizeEmail('  User@Example.COM  ')).toBe('user@example.com');
  });
});
