import { describe, it, expect } from 'vitest';
import { isValidPhone, getCountryCode } from '../../src/validators/phone';
import { validPhone, createValidInput } from '../fixtures/valid-inputs';
import { expectValid, expectNormalized } from '../helpers/validator-assertions';

describe('Phone Validator - Country', () => {
  it('should validate phone with country code', () => {
    const input = createValidInput('phone');
    const result = isValidPhone(input);
    expectValid(result, input);
    expect(result).toBe(true);
  });

  it('should detect US country code', () => {
    const result = getCountryCode('+15551234567');
    expectNormalized(result, 'US');
    expect(result).toBe('US');
  });

  it('should detect UK country code', () => {
    const result = getCountryCode('+447911123456');
    expectNormalized(result, 'UK');
    expect(result).toBe('UK');
  });
});
