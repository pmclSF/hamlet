describe('OrderProcessor', () => {
  it('should populate all order fields correctly', () => {
    const order = processOrder({ item: 'Widget', qty: 3, price: 9.99 });
    expect(order.id).toBeDefined();
    expect(order.item).toBe('Widget');
    expect(order.quantity).toBe(3);
    expect(order.unitPrice).toBeCloseTo(9.99);
    expect(order.total).toBeCloseTo(29.97);
    expect(order.status).toBe('pending');
    expect(order.createdAt).toBeInstanceOf(Date);
  });

  it('should apply discount to all line items', () => {
    const receipt = applyDiscount([
      { name: 'A', price: 100 },
      { name: 'B', price: 200 },
      { name: 'C', price: 50 },
    ], 0.1);
    expect(receipt[0].discountedPrice).toBeCloseTo(90);
    expect(receipt[1].discountedPrice).toBeCloseTo(180);
    expect(receipt[2].discountedPrice).toBeCloseTo(45);
    expect(receipt).toHaveLength(3);
  });

  it('should validate shipping address fields', () => {
    const addr = normalizeAddress({ street: '123 main st', city: 'springfield', zip: '62704' });
    expect(addr.street).toBe('123 Main St');
    expect(addr.city).toBe('Springfield');
    expect(addr.zip).toBe('62704');
    expect(addr.country).toBe('US');
  });
});
