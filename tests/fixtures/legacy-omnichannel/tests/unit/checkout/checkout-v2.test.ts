import { describe, it, expect } from 'vitest';
import { initiateCheckout, completeCheckout } from '../../../src/checkout/checkout';

describe('checkout v2', () => {
  it('should initiate', () => {
    expect(initiateCheckout('u1', 'cart_1').status).toBe('initiated');
  });
  it('should complete', () => {
    expect(completeCheckout('co_1', 'pay_token').status).toBe('completed');
  });
});
