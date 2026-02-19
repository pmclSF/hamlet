import { describe, it, expect, beforeEach, afterEach } from '@jest/globals';
import { CartService } from '../../src/cart-service.js';
import { ProductRepository } from '../../src/product-repository.js';

describe('CartService', () => {
  let cart;

  beforeEach(() => {
    cart = new CartService();
  });

  afterEach(() => {
    cart.clear();
  });

  it('should start with an empty cart', () => {
    expect(cart.itemCount()).toBe(0);
  });

  it('should add an item to the cart', () => {
    cart.addItem({ id: 'sku-001', name: 'Widget', price: 9.99 });
    expect(cart.itemCount()).toBe(1);
    expect(cart.getItems()).toEqual([
      { id: 'sku-001', name: 'Widget', price: 9.99, quantity: 1 },
    ]);
  });

  describe('with existing items', () => {
    beforeEach(() => {
      cart.addItem({ id: 'sku-001', name: 'Widget', price: 9.99 });
      cart.addItem({ id: 'sku-002', name: 'Gadget', price: 24.99 });
    });

    it('should calculate the total price', () => {
      expect(cart.totalPrice()).toBe(34.98);
    });

    it('should contain the added product names', () => {
      const names = cart.getItems().map((item) => item.name);
      expect(names).toContain('Widget');
      expect(names).toContain('Gadget');
    });

    it('should throw when removing a non-existent item', () => {
      expect(() => cart.removeItem('sku-999')).toThrow('Item not found');
    });

    it('should fetch product details asynchronously', async () => {
      const repo = new ProductRepository();
      const product = await repo.findById('sku-001');
      expect(product.name).toBe('Widget');
    });
  });
});
