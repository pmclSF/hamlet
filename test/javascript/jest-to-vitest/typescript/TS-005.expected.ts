import { describe, it, expect } from 'vitest';

interface Product {
  id: number;
  name: string;
  price: number;
}

type CartItem = {
  product: Product;
  quantity: number;
};

type Cart = CartItem[];

describe('Cart', () => {
  it('should calculate total', () => {
    const cart: Cart = [
      { product: { id: 1, name: 'Widget', price: 9.99 }, quantity: 2 },
      { product: { id: 2, name: 'Gadget', price: 24.99 }, quantity: 1 },
    ];
    const total: number = cart.reduce(
      (sum: number, item: CartItem) => sum + item.product.price * item.quantity,
      0,
    );
    expect(total).toBeCloseTo(44.97);
  });

  it('should handle empty cart', () => {
    const cart: Cart = [];
    expect(cart.length).toBe(0);
  });
});
