describe('mocks', () => {
  it('checks args', () => {
    const fn = jest.fn();
    fn('a', 'b');
    expect(fn.mock.calls[0]).toEqual(['a', 'b']);
  });
});
