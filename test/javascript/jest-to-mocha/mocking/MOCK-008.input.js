describe('test', () => {
  it('called with', () => {
    const fn = jest.fn();
    fn('a', 'b');
    expect(fn).toHaveBeenCalledWith('a', 'b');
  });
});
