describe('async', () => {
  it('uses async await', async () => {
    const val = await Promise.resolve(42);
    expect(val).toBe(42);
  });
});
