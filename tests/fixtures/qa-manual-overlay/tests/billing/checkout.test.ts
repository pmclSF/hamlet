import { describe, it, expect } from 'vitest';
import { calculateTotal } from '../../src/billing/checkout';

describe('Checkout', () => {
  it('should calculate total for cart items', () => {
    const items = [
      { sku: 'WIDGET-1', quantity: 2, priceInCents: 999 },
      { sku: 'GADGET-2', quantity: 1, priceInCents: 2499 },
    ];
    expect(calculateTotal(items)).toBe(4497);
  });

  it('should return 0 for empty cart', () => {
    expect(calculateTotal([])).toBe(0);
  });
});
