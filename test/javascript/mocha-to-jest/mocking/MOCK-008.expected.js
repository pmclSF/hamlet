describe('test', () => {
  it('calledWith', () => {
    const fn = jest.fn();
    fn('a', 'b');
    expect(fn).toHaveBeenCalledWith('a', 'b');
  });
});
