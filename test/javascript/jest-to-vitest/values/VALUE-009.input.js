describe('large numbers and BigInt', () => {
  it('should handle BigInt arithmetic', () => {
    const big = BigInt(Number.MAX_SAFE_INTEGER) + 1n;
    expect(big).toBe(9007199254740992n);
  });

  it('should compare BigInt values', () => {
    const a = 123456789012345678901234567890n;
    const b = 123456789012345678901234567890n;
    expect(a).toBe(b);
  });

  it('should handle MAX_SAFE_INTEGER boundary', () => {
    expect(Number.MAX_SAFE_INTEGER).toBe(9007199254740991);
    expect(Number.MIN_SAFE_INTEGER).toBe(-9007199254740991);
  });

  it('should handle very large regular numbers', () => {
    const large = 1e20;
    expect(large).toBe(100000000000000000000);
    expect(large).toBeGreaterThan(1e19);
  });
});
