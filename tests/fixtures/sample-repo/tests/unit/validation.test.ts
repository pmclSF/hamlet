import { describe, it, expect } from 'vitest';
import { validateEmail, validatePassword, sanitizeInput } from '../../src/utils/validation.js';

describe('validation', () => {
  describe('validateEmail', () => {
    it('should accept valid email', () => {
      expect(validateEmail('user@example.com')).toBe(true);
    });

    it('should reject invalid email', () => {
      expect(validateEmail('not-an-email')).toBe(false);
    });

    it('should reject empty string', () => {
      expect(validateEmail('')).toBe(false);
    });
  });

  describe('validatePassword', () => {
    it('should accept long password', () => {
      expect(validatePassword('longpassword')).toBe(true);
    });

    it('should reject short password', () => {
      expect(validatePassword('short')).toBe(false);
    });
  });

  describe('sanitizeInput', () => {
    it('should escape HTML characters', () => {
      expect(sanitizeInput('<script>')).not.toContain('<');
    });

    it('should pass through safe strings', () => {
      expect(sanitizeInput('hello world')).toBe('hello world');
    });
  });
});
