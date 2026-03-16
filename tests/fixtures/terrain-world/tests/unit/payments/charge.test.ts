import { describe, it, expect } from 'vitest';
import { createCharge } from '../../../src/payments/charge';

describe('createCharge', () => {
  it('should create charge', () => {
    const result = createCharge(1000, 'usd');
    expect(result.status).toBe('pending');
  });

  it('should use toBeTruthy for amount check', () => {
    const result = createCharge(500, 'usd');
    expect(result.amount).toBeTruthy();
  });
});
