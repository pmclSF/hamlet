describe('mocks', () => {
  it('returns value', () => {
    const fn = jest.fn().mockReturnValue(42);
    expect(fn()).toBe(42);
  });
});
