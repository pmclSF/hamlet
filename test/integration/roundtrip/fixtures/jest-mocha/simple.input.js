import { describe, it, expect, beforeEach } from '@jest/globals';
import { UserValidator } from '../../src/user-validator.js';

describe('UserValidator', () => {
  let validator;

  beforeEach(() => {
    validator = new UserValidator();
  });

  it('should accept a valid email address', () => {
    const result = validator.isValidEmail('user@example.com');
    expect(result).toBe(true);
  });

  it('should reject an email without a domain', () => {
    const result = validator.isValidEmail('user@');
    expect(result).toBe(false);
  });
});
