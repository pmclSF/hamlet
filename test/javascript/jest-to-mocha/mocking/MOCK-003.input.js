describe('test', () => {
  it('returns', () => {
    const fn = jest.fn().mockReturnValue(42);
    expect(fn()).toBe(42);
  });
});
