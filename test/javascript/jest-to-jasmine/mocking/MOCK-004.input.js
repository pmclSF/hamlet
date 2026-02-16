describe('mocks', () => {
  it('implements', () => {
    const fn = jest.fn().mockImplementation(x => x * 2);
    expect(fn(5)).toBe(10);
  });
});
