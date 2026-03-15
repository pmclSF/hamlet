import { describe, it, expect } from 'vitest';
import { isValidEmail, normalizeEmail } from '../../src/validators/email';
import { validEmail, createValidInput } from '../fixtures/valid-inputs';
import { expectValid, expectInvalid } from '../helpers/validator-assertions';

describe('Email Validator - Format', () => {
  it('should validate standard email format', () => {
    const input = createValidInput('email');
    const result = isValidEmail(input);
    expectValid(result, input);
    expect(result).toBe(true);
  });

  it('should reject email without at sign', () => {
    const result = isValidEmail('userexample.com');
    expectInvalid(result, 'userexample.com');
    expect(result).toBe(false);
  });

  it('should reject email without domain', () => {
    const result = isValidEmail('user@');
    expectInvalid(result, 'user@');
    expect(result).toBe(false);
  });
});
