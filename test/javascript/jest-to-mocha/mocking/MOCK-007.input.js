describe('test', () => {
  it('called times', () => {
    const fn = jest.fn();
    fn();
    fn();
    expect(fn).toHaveBeenCalledTimes(2);
  });
});
