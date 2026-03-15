import { describe, it, expect } from 'vitest';
import { isValidEmail, normalizeEmail } from '../../src/validators/email';
import { validEmail, createValidInput } from '../fixtures/valid-inputs';
import { expectValid, expectNormalized } from '../helpers/validator-assertions';

describe('Email Validator - Basic', () => {
  it('should validate correct email format', () => {
    const input = createValidInput('email');
    const result = isValidEmail(input);
    expectValid(result, input);
    expect(result).toBe(true);
  });

  it('should reject empty email', () => {
    expect(isValidEmail('')).toBe(false);
  });

  it('should normalize email to lowercase', () => {
    const result = normalizeEmail('USER@EXAMPLE.COM');
    expectNormalized(result, 'user@example.com');
    expect(result).toBe('user@example.com');
  });
});
