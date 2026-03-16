import { describe, it, expect } from 'vitest';
import { initiateCheckout, completeCheckout } from '../../../src/checkout/checkout';

describe('initiateCheckout', () => {
  it('should initiate checkout', () => {
    const r = initiateCheckout('u1', 'cart_1');
    expect(r.status).toBe('initiated');
  });
});

describe('completeCheckout', () => {
  it('should complete with valid payment', () => {
    expect(completeCheckout('co_1', 'pay_token').status).toBe('completed');
  });
  it('should fail with invalid payment', () => {
    expect(completeCheckout('co_1', 'bad').status).toBe('failed');
  });
});
