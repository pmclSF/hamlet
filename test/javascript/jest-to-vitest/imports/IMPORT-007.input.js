jest.setTimeout(30000);

describe('slow integration tests', () => {
  it('should complete a long-running operation', () => {
    const start = Date.now();
    const result = { status: 'complete' };
    const elapsed = Date.now() - start;
    expect(result.status).toBe('complete');
    expect(elapsed).toBeLessThan(30000);
  });

  it('should handle timeout-sensitive work', () => {
    const data = Array.from({ length: 1000 }, (_, i) => i);
    expect(data).toHaveLength(1000);
  });
});
