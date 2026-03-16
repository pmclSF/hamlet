import { describe, it, expect } from 'vitest';
import { initiatePurchase, completePurchase, validateReceipt } from '../../../src/mobile/purchase';

describe('initiatePurchase v2', () => {
  it('should initiate purchase on iOS', () => {
    const r = initiatePurchase('prod_1', 'ios');
    expect(r.status).toBe('initiated');
  });

  it('should initiate purchase on Android', () => {
    const r = initiatePurchase('prod_1', 'android');
    expect(r.status).toBe('initiated');
  });
});

describe('completePurchase v2', () => {
  it('should complete', () => {
    expect(completePurchase('pur_1').status).toBe('completed');
  });
});

describe('validateReceipt v2', () => {
  it('should validate', () => {
    expect(validateReceipt('data', 'ios').valid).toBe(true);
  });
});
