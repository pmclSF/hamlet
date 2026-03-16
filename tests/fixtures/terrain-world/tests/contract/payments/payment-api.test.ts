import { describe, it, expect } from 'vitest';
import { createCharge, captureCharge, voidCharge } from '../../../src/payments/charge';

describe('payments API contract', () => {
  it('createCharge returns required fields', () => {
    const result = createCharge(100, 'usd');
    expect(result).toHaveProperty('chargeId');
    expect(result).toHaveProperty('amount');
    expect(result).toHaveProperty('currency');
    expect(result).toHaveProperty('status');
  });

  it('captureCharge returns chargeId and status', () => {
    const result = captureCharge('ch_1');
    expect(result).toHaveProperty('chargeId');
    expect(result).toHaveProperty('status');
  });
});
