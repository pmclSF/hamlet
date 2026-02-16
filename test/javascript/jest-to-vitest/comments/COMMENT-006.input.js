describe('price calculator', () => {
  it('should calculate total with tax', () => {
    const price = 100;
    const taxRate = 0.08;
    const total = price + price * taxRate;

    expect(total).toBe(108);

    // expect(total).toBeCloseTo(108.00, 2);
    // expect(total).toBeGreaterThan(100);
  });

  it('should apply discount correctly', () => {
    const price = 200;
    const discount = 0.1;
    // const memberDiscount = 0.05;

    const discounted = price * (1 - discount);
    expect(discounted).toBe(180);

    // Old calculation before the refactor:
    // const oldTotal = price - (price * discount) - (price * memberDiscount);
    // expect(oldTotal).toBe(170);
  });
});
