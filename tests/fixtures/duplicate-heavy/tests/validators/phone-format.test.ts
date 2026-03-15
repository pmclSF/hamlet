import { describe, it, expect } from 'vitest';
import { isValidPhone, normalizePhone } from '../../src/validators/phone';
import { validPhone, createValidInput } from '../fixtures/valid-inputs';
import { expectValid, expectInvalid } from '../helpers/validator-assertions';

describe('Phone Validator - Format', () => {
  it('should validate standard phone format', () => {
    const input = createValidInput('phone');
    const result = isValidPhone(input);
    expectValid(result, input);
    expect(result).toBe(true);
  });

  it('should reject phone with letters', () => {
    const result = isValidPhone('+1-555-ABC-4567');
    expectInvalid(result, '+1-555-ABC-4567');
    expect(result).toBe(false);
  });

  it('should reject too-short phone', () => {
    const result = isValidPhone('123');
    expectInvalid(result, '123');
    expect(result).toBe(false);
  });
});
