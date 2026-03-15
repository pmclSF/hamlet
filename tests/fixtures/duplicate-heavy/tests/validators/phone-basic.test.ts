import { describe, it, expect } from 'vitest';
import { isValidPhone, normalizePhone } from '../../src/validators/phone';
import { validPhone, createValidInput } from '../fixtures/valid-inputs';
import { expectValid, expectNormalized } from '../helpers/validator-assertions';

describe('Phone Validator - Basic', () => {
  it('should validate correct phone format', () => {
    const input = createValidInput('phone');
    const result = isValidPhone(input);
    expectValid(result, input);
    expect(result).toBe(true);
  });

  it('should reject empty phone', () => {
    expect(isValidPhone('')).toBe(false);
  });

  it('should normalize phone by removing spaces', () => {
    const result = normalizePhone('+1 555 123 4567');
    expectNormalized(result, '+15551234567');
    expect(result).toBe('+15551234567');
  });
});
