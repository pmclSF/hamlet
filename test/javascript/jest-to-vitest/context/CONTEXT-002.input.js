describe('ShoppingCart', () => {
  let cart;

  beforeEach(() => {
    cart = { items: [], total: 0 };
  });

  describe('when empty', () => {
    it('should have no items', () => {
      expect(cart.items).toHaveLength(0);
    });

    it('should have zero total', () => {
      expect(cart.total).toBe(0);
    });
  });

  describe('when items are added', () => {
    let item;

    beforeEach(() => {
      item = { name: 'Widget', price: 9.99 };
      cart.items.push(item);
      cart.total += item.price;
    });

    it('should contain the item', () => {
      expect(cart.items).toContainEqual(item);
    });

    it('should update the total', () => {
      expect(cart.total).toBe(9.99);
    });
  });
});
