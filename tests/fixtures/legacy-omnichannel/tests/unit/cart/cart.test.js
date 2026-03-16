const { addToCart, removeFromCart, getCart } = require('../../../src/cart/cart');

describe('addToCart', () => {
  it('should add item to cart', () => {
    const result = addToCart('u1', 'prod_1', 2);
    expect(result.items.length).toBe(1);
  });
});

describe('removeFromCart', () => {
  it('should remove item', () => {
    expect(removeFromCart('cart_1', 'prod_1').removed).toBe('prod_1');
  });
});

describe('getCart', () => {
  it('should return cart', () => {
    expect(getCart('cart_1').cartId).toBe('cart_1');
  });
});
