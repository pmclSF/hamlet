describe('test', () => {
  it('async', async () => {
    const val = await Promise.resolve(42);
    expect(val).toBe(42);
  });
});
