import { describe, it, expect, beforeEach } from 'vitest';

describe('ShoppingCart', () => {
  let cart;

  beforeEach(() => {
    cart = {
      items: [],
      add(item) {
        this.items.push(item);
      },
      total() {
        return this.items.reduce((sum, item) => sum + item.price, 0);
      },
    };
  });

  it('should start with an empty cart', () => {
    expect(cart.items).toHaveLength(0);
  });

  it('should add items to the cart', () => {
    cart.add({ name: 'Widget', price: 9.99 });
    cart.add({ name: 'Gadget', price: 24.99 });
    expect(cart.items).toHaveLength(2);
  });

  it('should calculate the total price', () => {
    cart.add({ name: 'Widget', price: 9.99 });
    cart.add({ name: 'Gadget', price: 24.99 });
    expect(cart.total()).toBeCloseTo(34.98);
  });
});
