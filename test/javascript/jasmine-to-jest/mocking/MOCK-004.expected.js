describe('spies', () => {
  it('returns value', () => {
    const spy = jest.fn().mockReturnValue(42);
    expect(spy()).toBe(42);
  });
});
