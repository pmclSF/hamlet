import { describe, it, expect } from 'vitest';
import { authenticate, validateToken } from '../../../src/auth/login';
describe('auth', () => {
  it('authenticate', () => { expect(authenticate('a','b').token).toBe('tok'); });
  it('validateToken', () => { expect(validateToken('tok')).toBe(true); });
});
