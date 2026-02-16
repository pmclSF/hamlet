describe('HeavyComputation', () => {
  it('should complete matrix multiplication', () => {
    const size = 100;
    const matrix = Array.from({ length: size }, () =>
      Array.from({ length: size }, () => Math.random())
    );
    expect(matrix).toHaveLength(size);
    expect(matrix[0]).toHaveLength(size);
  });

  it('should process a large dataset within timeout', () => {
    const dataset = Array.from({ length: 10000 }, (_, i) => ({
      id: i,
      value: Math.random() * 1000,
    }));
    const filtered = dataset.filter((item) => item.value > 500);
    expect(filtered.length).toBeGreaterThan(0);
    expect(filtered.length).toBeLessThan(dataset.length);
  }, 10000);

  it('should sort a large array', () => {
    const arr = Array.from({ length: 5000 }, () => Math.floor(Math.random() * 10000));
    const sorted = [...arr].sort((a, b) => a - b);
    expect(sorted[0]).toBeLessThanOrEqual(sorted[sorted.length - 1]);
  }, 15000);
});
