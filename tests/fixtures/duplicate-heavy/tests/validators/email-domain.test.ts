import { describe, it, expect } from 'vitest';
import { isValidEmail, getDomain } from '../../src/validators/email';
import { validEmail, createValidInput } from '../fixtures/valid-inputs';
import { expectValid, expectNormalized } from '../helpers/validator-assertions';

describe('Email Validator - Domain', () => {
  it('should validate email with valid domain', () => {
    const input = createValidInput('email');
    const result = isValidEmail(input);
    expectValid(result, input);
    expect(result).toBe(true);
  });

  it('should extract domain from email', () => {
    const result = getDomain(validEmail);
    expectNormalized(result, 'example.com');
    expect(result).toBe('example.com');
  });

  it('should handle email with subdomain', () => {
    expect(isValidEmail('user@mail.example.com')).toBe(true);
    expect(getDomain('user@mail.example.com')).toBe('mail.example.com');
  });
});
