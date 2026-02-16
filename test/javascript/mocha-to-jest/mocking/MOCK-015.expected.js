describe('test', () => {
  it('callCount', () => {
    const fn = jest.fn();
    fn();
    fn();
    fn();
    expect(fn).toHaveBeenCalledTimes(3);
  });
});
