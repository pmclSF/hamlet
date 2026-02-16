describe('test', () => {
  it('stub', () => {
    const fn = jest.fn();
    fn();
    expect(fn).toHaveBeenCalled();
  });
});
