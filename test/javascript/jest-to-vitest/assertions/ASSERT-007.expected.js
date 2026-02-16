import { describe, it, expect } from 'vitest';

describe('ShoppingCart', () => {
  it('should contain the added item by reference', () => {
    const cart = new Cart();
    cart.add('apple');
    expect(cart.items()).toContain('apple');
  });

  it('should contain an object with matching properties', () => {
    const cart = new Cart();
    cart.add({ id: 1, name: 'Widget', price: 9.99 });
    expect(cart.items()).toContainEqual({ id: 1, name: 'Widget', price: 9.99 });
  });

  it('should list all supported currencies', () => {
    const currencies = getSupportedCurrencies();
    expect(currencies).toContain('USD');
    expect(currencies).toContain('EUR');
    expect(currencies).toContain('GBP');
  });
});
