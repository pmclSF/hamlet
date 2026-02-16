describe('mocks', () => {
  it('checks call count', () => {
    const fn = jest.fn();
    fn();
    fn();
    expect(fn.mock.calls.length).toBe(2);
  });
});
