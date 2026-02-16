describe('mocks', () => {
  it('resolves', async () => {
    const fn = jest.fn().mockResolvedValue('data');
    const result = await fn();
    expect(result).toBe('data');
  });
});
