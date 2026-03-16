import { describe, it, expect } from 'vitest';
import { initiatePurchase, completePurchase, validateReceipt } from '../../../src/mobile/purchase';

describe('initiatePurchase', () => {
  it('should initiate iOS purchase', () => {
    const result = initiatePurchase('prod_1', 'ios');
    expect(result.status).toBe('initiated');
    expect(result.platform).toBe('ios');
  });

  it('should initiate Android purchase', () => {
    const result = initiatePurchase('prod_1', 'android');
    expect(result.status).toBe('initiated');
    expect(result.platform).toBe('android');
  });
});

describe('completePurchase', () => {
  it('should complete purchase', () => {
    expect(completePurchase('pur_1').status).toBe('completed');
  });
});

describe('validateReceipt', () => {
  it('should validate iOS receipt', () => {
    expect(validateReceipt('receipt_data', 'ios').valid).toBe(true);
  });

  it('should validate Android receipt', () => {
    expect(validateReceipt('receipt_data', 'android').valid).toBe(true);
  });
});
