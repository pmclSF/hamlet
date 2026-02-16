describe('Infinity assertions', () => {
  it('should detect positive Infinity', () => {
    expect(1 / 0).toBe(Infinity);
    expect(Number.POSITIVE_INFINITY).toBe(Infinity);
  });

  it('should detect negative Infinity', () => {
    expect(-1 / 0).toBe(-Infinity);
    expect(Number.NEGATIVE_INFINITY).toBe(-Infinity);
  });

  it('should distinguish Infinity from finite numbers', () => {
    expect(Number.isFinite(42)).toBe(true);
    expect(Number.isFinite(Infinity)).toBe(false);
    expect(Number.isFinite(-Infinity)).toBe(false);
  });

  it('should handle Infinity in comparisons', () => {
    expect(Infinity).toBeGreaterThan(Number.MAX_VALUE);
    expect(-Infinity).toBeLessThan(-Number.MAX_VALUE);
  });
});
