describe('test', () => {
  it('clear', () => {
    const fn = jest.fn();
    fn();
    fn.mockClear();
    expect(fn).not.toHaveBeenCalled();
  });
});
