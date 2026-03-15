import { describe, it, expect } from 'vitest';
import { hashPassword, generateToken } from '../../src/utils/crypto.js';

describe('crypto', () => {
  describe('hashPassword', () => {
    it('should hash the password', () => {
      const hash = hashPassword('mypassword');
      expect(hash).toBeDefined();
      expect(hash).not.toBe('mypassword');
    });

    it('should produce consistent hashes', () => {
      expect(hashPassword('same')).toBe(hashPassword('same'));
    });

    it('should produce different hashes for different inputs', () => {
      expect(hashPassword('one')).not.toBe(hashPassword('two'));
    });
  });

  describe('generateToken', () => {
    it('should produce a token string', () => {
      const token = generateToken();
      expect(token).toBeDefined();
      expect(typeof token).toBe('string');
    });
  });
});
